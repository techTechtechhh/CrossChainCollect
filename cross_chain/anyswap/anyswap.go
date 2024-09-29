package anyswap

import (
	"app/model"
	"app/svc"
	"app/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
	"math/big"
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

var _ model.EventCollector = &Anyswap{}

type Anyswap struct {
	svc *svc.ServiceContext
}

func NewAnyswapCollector(svc *svc.ServiceContext) *Anyswap {
	return &Anyswap{
		svc: svc,
	}
}

func (a *Anyswap) Name() string {
	return "Anyswap"
}

func (a *Anyswap) Contracts(chain string) map[string]string {
	return make(map[string]string)
}

func (a *Anyswap) Topics0(chain string) []string {
	return []string{LogAnySwapIn, LogAnySwapOut, LogAnySwapOut2}
}

func (a *Anyswap) SrcTopics0() []string {
	return []string{
		LogAnySwapOut, LogAnySwapOut2,
	}
}
func (a *Anyswap) Extract(chain string, events model.Events) model.Results {
	ret := make(model.Results, 0)
	for _, e := range events {
		if _, ok := NonAnyswapContracts[e.Address]; ok {
			continue
		}
		res := &model.Result{
			Chain:    chain,
			Number:   e.Number,
			Ts:       e.Ts.UTC(),
			Index:    e.Index,
			Hash:     e.Hash,
			ActionId: e.Id,
			Project:  a.Name(),
			Contract: e.Address,
		}
		switch e.Topics[0] {
		case LogAnySwapIn:
			if len(e.Topics) < 4 || len(e.Data) < 2+3*64 {
				continue
			}
			res.Direction = model.InDirection
			fromChainId, _ := new(big.Int).SetString(e.Data[2+64:2+128], 16)
			res.FromChainId = (*model.BigInt)(fromChainId)
			toChainId, _ := new(big.Int).SetString(e.Data[2+128:2+192], 16)
			res.ToChainId = (*model.BigInt)(toChainId)
			res.ToAddress.Scan("0x" + e.Topics[3][26:])
			res.Token = "0x" + e.Topics[2][26:]
			amount, _ := new(big.Int).SetString(e.Data[2:2+64], 16)
			res.Amount = (*model.BigInt)(amount)
			d := &Detail{
				SrcTxHash: e.Topics[1],
			}
			detail, err := json.Marshal(d)
			if err == nil {
				res.Detail = detail
			}
			res.MatchTag = e.Topics[1]

		case LogAnySwapOut:
			if len(e.Topics) < 4 || len(e.Data) < 2+3*64 {
				continue
			}
			res.Direction = model.OutDirection
			fromChainId, _ := new(big.Int).SetString(e.Data[2+64:2+128], 16)
			res.FromChainId = (*model.BigInt)(fromChainId)
			res.FromAddress.Scan("0x" + e.Topics[2][26:])
			toChainId, _ := new(big.Int).SetString(e.Data[2+128:2+192], 16)
			res.ToChainId = (*model.BigInt)(toChainId)
			res.ToAddress.Scan("0x" + e.Topics[3][26:])
			res.Token = "0x" + e.Topics[1][26:]
			amount, _ := new(big.Int).SetString(e.Data[2:2+64], 16)
			res.Amount = (*model.BigInt)(amount)
			res.MatchTag = e.Hash
		case LogAnySwapOut2:
			if len(e.Topics) < 3 {
				continue
			}
			res.Direction = model.OutDirection
			res.Token = "0x" + e.Topics[1][26:]
			res.FromAddress.Scan("0x" + e.Topics[2][26:])
			a, err := abi.JSON(bytes.NewBufferString(RouterV6["avalanche"]))

			ev, err := a.EventByID(common.HexToHash(LogAnySwapOut2))
			if err != nil {
				log.Error("Anyswap Exact() can't decode LogAnySwapOut2", "Chain", chain, "Hash", e.Hash)
			}
			ss, err := ev.Inputs.Unpack(hexutil.MustDecode(e.Data))
			res.ToAddress.Scan(strings.ToLower(ss[0].(string)))
			res.Amount = (*model.BigInt)(ss[1].(*big.Int))
			res.FromChainId = (*model.BigInt)(ss[2].(*big.Int))
			res.ToChainId = (*model.BigInt)(ss[3].(*big.Int))
			res.MatchTag = e.Hash

		}
		d := &Detail{
			SrcTxHash: res.MatchTag,
		}
		detail, err := json.Marshal(d)
		if err == nil {
			res.Detail = detail
		}
		res.MatchTag = updateAnyswapMatchTag(res.MatchTag)
		ret = append(ret, res)
	}
	return ret
}

func (a *Anyswap) GetUnderlying(chain, anyToken string) (string, error) {
	p := a.svc.Providers.Get(chain)
	if p == nil {
		return "", fmt.Errorf("providers does not support %v", chain)
	}
	raw, err := p.ContinueCall("", anyToken, Underlying, nil, nil)
	if err != nil {
		return "", err
	}
	return strings.ToLower(common.BytesToAddress(raw).Hex()), nil
}

func updateAnyswapMatchTag(matchTag string) string {
	var isStringAlphabetic = regexp.MustCompile(`^[0-9]+$`).MatchString
	// 若包含字母则返回false，不包含字母则返回true

	if ert := isStringAlphabetic(matchTag[2:]); !ert { //是更新前的形式，即srcTxHash，需要进一步处理
		var swapIDHash common.Hash
		if utils.IsHex(matchTag) {
			swapIDHash = common.HexToHash(matchTag)
		} else {
			swapIDHash = common.BytesToHash([]byte(matchTag))
		}
		matchTag = swapIDHash.String()
	}
	return matchTag
}
