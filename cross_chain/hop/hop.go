package hop

import (
	"app/model"
	"app/utils"
	"encoding/json"
	"math/big"
)

var _ model.EventCollector = &Hop{}

type Hop struct {
}

func NewHopCollector() *Hop {
	return &Hop{}
}

func (a *Hop) Name() string {
	return "Hop"
}

func (a *Hop) Contracts(chain string) map[string]string {
	return make(map[string]string)
	/*if _, ok := hopContracts[chain]; !ok {
		return nil
	}
	return hopContracts[chain]*/
}

func (a *Hop) Topics0(chain string) []string {
	return []string{TransferSent, TransferSentToL2,
		WithdrawalBonded, TransferFromL1Completed}
}

func (a *Hop) SrcTopics0() []string {
	return []string{
		TransferSent, TransferSentToL2,
	}
}

func (a *Hop) Extract(chain string, events model.Events) model.Results {
	ret := make(model.Results, 0)

	for i := 0; i < len(events); i++ {
		e := events[i]

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

		d := &Detail{}
		ddl := &big.Int{}
		minDy := &big.Int{}
		relayer := ""
		transferID := ""

		switch e.Topics[0] {
		case TransferSentToL2:
			res.Direction = model.OutDirection
			toChainId, _ := new(big.Int).SetString(e.Topics[1][2:], 16)
			res.ToChainId = (*model.BigInt)(toChainId)
			res.ToAddress.Scan("0x" + e.Topics[2][26:])
			amount, _ := new(big.Int).SetString(e.Data[2:66], 16)
			res.Amount = (*model.BigInt)(amount)
			minDy, _ = new(big.Int).SetString(e.Data[66:130], 16)
			ddl, _ = new(big.Int).SetString(e.Data[130:194], 16)
			d.MinDy = *minDy
			relayer = "0x" + e.Topics[3][26:]

		case TransferSent:
			res.Direction = model.OutDirection
			transferID = e.Topics[1]
			toChainId, _ := new(big.Int).SetString(e.Topics[2][2:], 16)
			res.ToChainId = (*model.BigInt)(toChainId)
			res.ToAddress.Scan("0x" + e.Topics[3][26:])
			amount, _ := new(big.Int).SetString(e.Data[2:66], 16)
			res.Amount = (*model.BigInt)(amount)
			minDy, _ = new(big.Int).SetString(e.Data[258:322], 16)
			d.MinDy = *minDy
			ddl, _ = new(big.Int).SetString(e.Data[322:], 16)

		case TransferFromL1Completed:
			res.Direction = model.InDirection
			res.ToAddress.Scan("0x" + e.Topics[1][26:])
			relayer = "0x" + e.Topics[2][26:]
			amount, _ := new(big.Int).SetString(e.Data[2:66], 16)
			res.Amount = (*model.BigInt)(amount)
			minDy, _ = new(big.Int).SetString(e.Data[66:130], 16)
			d.MinDy = *minDy
			ddl, _ = new(big.Int).SetString(e.Data[130:194], 16)

		case WithdrawalBonded:
			res.Direction = model.InDirection
			transferID = e.Topics[1]
			amount, _ := new(big.Int).SetString(e.Data[2:], 16)
			res.Amount = (*model.BigInt)(amount)
		}

		d.DDL = *ddl
		d.Relayer = relayer
		d.TransferID = transferID
		res.Token = hopToken[chain][e.Address]

		if len(transferID) != 0 {
			res.MatchTag = transferID
		} else {
			res.MatchTag = ddl.String() + res.ToAddress.String + d.MinDy.String()
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
