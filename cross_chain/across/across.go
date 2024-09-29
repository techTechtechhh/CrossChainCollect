package across

import (
	"app/model"
	"app/utils"
	"encoding/json"
	"fmt"
	"math/big"
)

var _ model.EventCollector = &Across{}

type Across struct {
}

func NewAcrossCollector() *Across {
	return &Across{}
}

func (a *Across) Name() string {
	return "Across"
}

func (a *Across) Contracts(chain string) map[string]string {
	return make(map[string]string)
}

func (a *Across) Topics0(chain string) []string {
	return []string{FundsDeposited, FilledRelay, FundsDeposited2, FilledRelay2}
}

func (a *Across) SrcTopics0() []string {
	return []string{
		FundsDeposited, FundsDeposited2,
	}
}

func (a *Across) Extract(chain string, events model.Events) model.Results {
	ret := make(model.Results, 0)
	len_out := 386
	len_in := 834

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

		switch e.Topics[0] {
		case FundsDeposited:
			if len(e.Topics) < 4 || len(e.Data) < len_out {
				continue
			}
			res.Direction = model.OutDirection
			fromChainId, _ := new(big.Int).SetString(e.Data[2+64:2+128], 16)
			res.FromChainId = (*model.BigInt)(fromChainId)
			toChainId, _ := new(big.Int).SetString(e.Data[2+128:2+192], 16)
			res.ToChainId = (*model.BigInt)(toChainId)
			res.ToAddress.Scan("0x" + e.Data[len_out-64+24:])
			res.FromAddress.Scan("0x" + e.Topics[3][26:])
			res.Token = "0x" + e.Topics[2][26:]
			amount, _ := new(big.Int).SetString(e.Data[2:2+64], 16)
			res.Amount = (*model.BigInt)(amount)

			depositId := e.Topics[1]
			d := &Detail{
				DepositId: depositId,
			}
			detail, err := json.Marshal(d)
			if err == nil {
				res.Detail = detail
			}
			res.MatchTag = d.DepositId

		case FilledRelay:
			if len(e.Topics) < 3 || len(e.Data) < len_in {
				continue
			}
			amount, _ := new(big.Int).SetString(e.Data[2:66], 16)
			fillAmoun, _ := new(big.Int).SetString(e.Data[66:130], 16)
			if amount.Cmp(fillAmoun) > 0 {
				continue
			} else if amount.Cmp(fillAmoun) < 0 {
				utils.SendMail("Across amount < fill_amount", fmt.Sprintf("Hash: %s, Chain: %s, amount: %s, filledAmount: %s", e.Hash, chain, amount.String(), fillAmoun.String()))
				continue
			}
			res.Direction = model.InDirection
			relayer := "0x" + e.Topics[1][26:]
			fromChainId, _ := new(big.Int).SetString(e.Data[2+64*4:2+64*5], 16)
			res.FromChainId = (*model.BigInt)(fromChainId)
			res.FromAddress.Scan("0x" + e.Topics[2][26:])

			toChainId, _ := new(big.Int).SetString(e.Data[2+64*5:2+64*6], 16)
			res.ToChainId = (*model.BigInt)(toChainId)
			depositId := "0x" + e.Data[len_in-64*4:len_in-64*3]
			res.Token = "0x" + e.Data[len_in-64*3+24:len_in-128]
			res.ToAddress.Scan("0x" + e.Data[len_in-64*2+24:len_in-64])
			res.Amount = (*model.BigInt)(amount)
			d := &Detail{
				DepositId: depositId,
				Relayer:   relayer,
			}
			detail, err := json.Marshal(d)
			if err == nil {
				res.Detail = detail
			}
			res.MatchTag = d.DepositId

		case FundsDeposited2:
			if len(e.Topics) < 4 {
				continue
			}
			res.Direction = model.OutDirection
			fromChainId, _ := new(big.Int).SetString(e.Data[2+64:2+128], 16)
			res.FromChainId = (*model.BigInt)(fromChainId)
			toChainId, _ := new(big.Int).SetString(e.Topics[1][2:], 16)
			res.ToChainId = (*model.BigInt)(toChainId)
			res.ToAddress.Scan("0x" + e.Data[26+320:66+320])
			res.FromAddress.Scan("0x" + e.Topics[3][26:66])
			res.Token = "0x" + e.Data[26+256:66+256]
			amount, _ := new(big.Int).SetString(e.Data[2:2+64], 16)
			res.Amount = (*model.BigInt)(amount)

			depositId := e.Topics[2]
			d := &Detail{
				DepositId: depositId,
			}
			detail, err := json.Marshal(d)
			if err == nil {
				res.Detail = detail
			}
			res.MatchTag = d.DepositId

		case FilledRelay2:
			if len(e.Topics) < 4 {
				continue
			}
			res.Direction = model.InDirection
			relayer := "0x" + e.Data[26+8*64:66+8*64]
			res.FromAddress.Scan("0x" + e.Topics[3][26:])
			res.ToAddress.Scan("0x" + e.Data[26+9*64:66+9*64])

			fromChainId, _ := new(big.Int).SetString(e.Topics[1][2:], 16)
			res.FromChainId = (*model.BigInt)(fromChainId)
			toChainId, _ := new(big.Int).SetString(e.Data[2+256:66+256], 16)
			res.ToChainId = (*model.BigInt)(toChainId)
			depositId := e.Topics[2]
			res.Token = "0x" + e.Data[26+448:66+448]
			amount, _ := new(big.Int).SetString(e.Data[2:66], 16)
			res.Amount = (*model.BigInt)(amount)
			d := &Detail{
				DepositId: depositId,
				Relayer:   relayer,
			}
			detail, err := json.Marshal(d)
			if err == nil {
				res.Detail = detail
			}
			res.MatchTag = d.DepositId
		}
		ret = append(ret, res)
	}
	return ret
}
