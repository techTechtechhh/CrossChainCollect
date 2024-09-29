package celer_bridge

import (
	"app/model"
	"app/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
	"math/big"
	"strings"
)

var _ model.EventCollector = &CBridge{}

type CBridge struct {
}

func NewCBridgeCollector() *CBridge {
	return &CBridge{}
}

func (a *CBridge) Name() string {
	return "CBridge"
}

func (a *CBridge) Contracts(chain string) map[string]string {
	/*if _, ok := CBridgeContracts[chain]; !ok {
		return nil
	}
	return CBridgeContracts[chain]*/
	return make(map[string]string)
}

func (a *CBridge) Topics0(chain string) []string {
	return []string{Burn_2, Send, Deposited_1, Deposited_2, Burn_1,
		Mint, Relay, Withdrawn}
}

func (a *CBridge) SrcTopics0() []string {
	return []string{
		Burn_1, Burn_2, Deposited_1, Deposited_2, Send,
	}
}

func (a *CBridge) Extract(chain string, events model.Events) model.Results {
	ret := make(model.Results, 0)

	for _, e := range events {
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

		d := &Detail{
			TxId: "0",
		}

		if e.Topics[0] == Burn_2 {
			res.Direction = model.OutDirection
			res.MatchTag = e.Data[:66]
			res.Token = "0x" + e.Data[90:130]
			res.FromAddress.Scan("0x" + e.Data[154:194])
			amount, _ := new(big.Int).SetString(e.Data[194:258], 16)
			res.Amount = (*model.BigInt)(amount)
			d.TxId = res.MatchTag
			nonce, _ := new(big.Int).SetString(e.Data[386:450], 16)
			toChainId, _ := new(big.Int).SetString(e.Data[258:322], 16)
			res.ToChainId = (*model.BigInt)(toChainId)
			res.ToAddress.Scan("0x" + e.Data[322:386])
			d.Nonce = *nonce

		} else if e.Topics[0] == Send {
			res.Direction = model.OutDirection
			res.MatchTag = e.Data[:66]
			res.FromAddress.Scan("0x" + e.Data[90:130])
			res.ToAddress.Scan("0x" + e.Data[154:194])
			res.Token = "0x" + e.Data[218:258]
			amount, _ := new(big.Int).SetString(e.Data[258:322], 16)
			res.Amount = (*model.BigInt)(amount)
			toChainId, _ := new(big.Int).SetString(e.Data[322:386], 16)
			res.ToChainId = (*model.BigInt)(toChainId)
			nonce, _ := new(big.Int).SetString(e.Data[386:450], 16)
			maxSlippage, _ := new(big.Int).SetString(e.Data[450:514], 16)
			d.TxId = res.MatchTag
			d.Nonce = *nonce
			d.MaxSlippage = *maxSlippage

		} else if e.Topics[0] == Deposited_1 || e.Topics[0] == Deposited_2 {
			res.Direction = model.OutDirection
			res.MatchTag = e.Data[:66]
			res.FromAddress.Scan("0x" + e.Data[90:130])
			res.Token = "0x" + e.Data[154:194]
			amount, _ := new(big.Int).SetString(e.Data[194:258], 16)
			res.Amount = (*model.BigInt)(amount)
			toChainId, _ := new(big.Int).SetString(e.Data[258:322], 16)
			res.ToChainId = (*model.BigInt)(toChainId)
			res.ToAddress.Scan("0x" + e.Data[346:386])
			d.TxId = res.MatchTag

			if e.Topics[0] == Deposited_2 {
				nonce, _ := new(big.Int).SetString(e.Data[386:], 16)
				d.Nonce = *nonce
			}

		} else if e.Topics[0] == Mint {
			res.Direction = model.InDirection
			mintId := e.Data[:66]
			res.Token = "0x" + e.Data[90:130]
			res.ToAddress.Scan("0x" + e.Data[154:194])
			amount, _ := new(big.Int).SetString(e.Data[194:258], 16)
			res.Amount = (*model.BigInt)(amount)
			fromChainId, _ := new(big.Int).SetString(e.Data[258:322], 16)
			res.FromChainId = (*model.BigInt)(fromChainId)
			res.MatchTag = "0x" + e.Data[322:386]
			res.FromAddress.Scan("0x" + e.Data[410:])
			d.TxId = mintId

		} else if e.Topics[0] == Relay {
			res.Direction = model.InDirection
			tsfId := e.Data[:66]
			res.FromAddress.Scan("0x" + e.Data[90:130])
			res.ToAddress.Scan("0x" + e.Data[154:194])
			res.Token = "0x" + e.Data[218:258]
			amount, _ := new(big.Int).SetString(e.Data[258:322], 16)
			res.Amount = (*model.BigInt)(amount)
			fromChainId, _ := new(big.Int).SetString(e.Data[322:386], 16)
			res.FromChainId = (*model.BigInt)(fromChainId)
			res.MatchTag = "0x" + e.Data[386:450]
			d.TxId = tsfId

		} else if e.Topics[0] == Withdrawn {
			res.Direction = model.InDirection
			wdId := e.Data[:66]
			res.ToAddress.Scan("0x" + e.Data[90:130])
			res.Token = "0x" + e.Data[154:194]
			amount, _ := new(big.Int).SetString(e.Data[194:258], 16)
			res.Amount = (*model.BigInt)(amount)
			fromChainId, _ := new(big.Int).SetString(e.Data[258:322], 16)
			res.FromChainId = (*model.BigInt)(fromChainId)
			res.MatchTag = "0x" + e.Data[322:386]
			d.TxId = wdId
			res.FromAddress.Scan("0x" + e.Data[410:])

		} else if e.Topics[0] == Burn_1 {
			res.Direction = model.OutDirection
			a, err := abi.JSON(bytes.NewBufferString(PeggedContract["avalanche"]))
			ev, err := a.EventByID(common.HexToHash(Burn_1))
			if err != nil {
				log.Error("Poly Exact() can't decode CrossChainEvent", "Chain", chain, "Hash", e.Hash)
			}
			ss, err := ev.Inputs.Unpack(hexutil.MustDecode(e.Data))
			res.MatchTag = strings.ToLower(fmt.Sprintf("0x%x", ss[0].([32]uint8)))
			res.Token = strings.ToLower(fmt.Sprintf("%v", ss[1].(common.Address)))
			res.FromAddress.Scan(strings.ToLower(fmt.Sprintf("%v", ss[2].(common.Address))))
			res.Amount = (*model.BigInt)(ss[3].(*big.Int))
			res.ToAddress.Scan(strings.ToLower(fmt.Sprintf("%v", ss[4].(common.Address))))

		} else {
			continue
		}

		if res.Direction == model.InDirection {
			toChainId := new(big.Int).Set(utils.GetChainId(chain))
			res.ToChainId = (*model.BigInt)(toChainId)
		} else if res.Direction == model.OutDirection {
			fromChainId := new(big.Int).Set(utils.GetChainId(chain))
			res.FromChainId = (*model.BigInt)(fromChainId)
		}

		detail, err := json.Marshal(d)
		if err == nil {
			res.Detail = detail
		}

		ret = append(ret, res)
	}
	return ret
}
