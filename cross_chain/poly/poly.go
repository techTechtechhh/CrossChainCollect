package poly

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

var _ model.EventCollector = &Poly{}

type Poly struct{}

var expoitor = strings.ToLower("E0AFADAD1D93704761C8550F21A53DE3468BA599")

func NewPolyCollector() *Poly {
	return &Poly{}
}

func (a *Poly) Name() string {
	return "PolyNetwork"
}

func (a *Poly) Contracts(chain string) map[string]string {
	return make(map[string]string)
	/*if _, ok := PolyContracts[chain]; !ok {
		return nil
	}
	return PolyContracts[chain]*/
}

func (a *Poly) Topics0(chain string) []string {
	return []string{CrossChainEvent, VerifyHeaderAndExecuteTxEvent, LockEvent, UnLockEvent, UnLockEvent_Switcheo}
}

func (a *Poly) SrcTopics0() []string {
	return []string{CrossChainEvent, LockEvent, LockEvent_Switcheo}
}

func (a *Poly) Extract(chain string, events model.Events) model.Results {
	retUnlock := make(model.Results, 0)
	retVerify := make(model.Results, 0)
	retLock := make(model.Results, 0)
	retCross := make(model.Results, 0)

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
		var d = &Detail{}
		switch e.Topics[0] {
		case CrossChainEvent:
			if len(e.Topics) < 2 {
				continue
			}
			a, err := abi.JSON(bytes.NewBufferString(EthCrossChainManager["bsc"]))
			ev, err := a.EventByID(common.HexToHash(CrossChainEvent))
			if err != nil {
				log.Error("Poly Exact() can't decode CrossChainEvent", "Chain", chain, "Hash", e.Hash)
			}
			ss, err := ev.Inputs.Unpack(hexutil.MustDecode(e.Data))
			res.MatchTag = strings.TrimLeft(fmt.Sprintf("%x", ss[0].([]uint8)), "0")
			res.ToChainId = switchChainId(ss[2])
			d.ToContract = strings.ToLower(fmt.Sprintf("0x%x", ss[3].([]uint8)))
			retCross = append(retCross, res)
		case VerifyHeaderAndExecuteTxEvent:
			a, err := abi.JSON(bytes.NewBufferString(EthCrossChainManager["bsc"]))
			ev, err := a.EventByID(common.HexToHash(VerifyHeaderAndExecuteTxEvent))
			if err != nil {
				log.Error("Poly Exact() can't decode VerifyHeader", "Chain", chain, "Hash", e.Hash)
			}
			ss, err := ev.Inputs.Unpack(hexutil.MustDecode(e.Data))
			res.FromChainId = switchChainId(ss[0])
			d.CrossChainTxHash = strings.ToLower(fmt.Sprintf("%x", ss[2].([]uint8)))
			d.ToContract = strings.ToLower(fmt.Sprintf("0x%x", ss[1].([]uint8)))
			res.MatchTag = strings.TrimLeft(fmt.Sprintf("%x", ss[3].([]uint8)), "0")
			toChainId := new(big.Int).Set(utils.GetChainId(chain))
			res.ToChainId = (*model.BigInt)(toChainId)
			retVerify = append(retVerify, res)
		case LockEvent:
			res.Direction = model.OutDirection
			a, err := abi.JSON(bytes.NewBufferString(LockManager["eth"]))
			ev, err := a.EventByID(common.HexToHash(LockEvent))
			if err != nil {
				log.Error("Poly Exact() can't decode LockEvent", "Chain", chain, "Hash", e.Hash)
			}
			ss, err := ev.Inputs.Unpack(hexutil.MustDecode(e.Data))
			res.Token = strings.ToLower(fmt.Sprintf("%v", ss[0].(common.Address)))
			res.FromAddress.Scan(strings.ToLower(ss[1].(common.Address).Hex()))
			res.ToChainId = switchChainId(ss[2])
			d.ToAsset = strings.ToLower(fmt.Sprintf("0x%x", ss[3].([]uint8)))
			res.ToAddress.Scan(fmt.Sprintf("0x%x", ss[4].([]uint8)))
			res.Amount = (*model.BigInt)(ss[5].(*big.Int))
			fromChainId := new(big.Int).Set(utils.GetChainId(chain))
			res.FromChainId = (*model.BigInt)(fromChainId)
			retLock = append(retLock, res)
		case UnLockEvent:
			res.Direction = model.InDirection
			a, err := abi.JSON(bytes.NewBufferString(LockManager["eth"]))
			ev, err := a.EventByID(common.HexToHash(UnLockEvent))
			if err != nil {
				log.Error("Poly Exact() can't decode LockEvent", "Chain", chain, "Hash", e.Hash)
			}
			ss, err := ev.Inputs.Unpack(hexutil.MustDecode(e.Data))
			res.ToAddress.Scan(strings.ToLower(ss[1].(common.Address).Hex()))
			res.Token = strings.ToLower(ss[0].(common.Address).Hex())
			res.Amount = (*model.BigInt)(ss[2].(*big.Int))
			retUnlock = append(retUnlock, res)
		case UnLockEvent_Switcheo:
			res.Direction = model.InDirection
			if len(e.Topics) == 3 {
				res.Token = "0x" + e.Topics[1][26:]
				res.ToAddress.Scan("0x" + e.Topics[2][26:])
				amount, _ := new(big.Int).SetString(e.Data[2:2+64], 16)
				res.Amount = (*model.BigInt)(amount)
			} else if len(e.Topics) == 1 {
				res.Token = "0x" + e.Data[26:66]
				res.ToAddress.Scan("0x" + e.Data[90:130])
				amount, _ := new(big.Int).SetString(e.Data[130:194], 16)
				res.Amount = (*model.BigInt)(amount)
			}
			retUnlock = append(retUnlock, res)
		case LockEvent_Switcheo:
			res.Direction = model.OutDirection
			if len(e.Topics) == 4 {
				res.Token = "0x" + e.Topics[1][26:]
				res.FromAddress.Scan("0x" + e.Topics[2][26:])
				res.ToChainId = switchChainId(e.Topics[3][2:])
				//res.ToChainId = (*model.BigInt)(toChainId)
				a, err := abi.JSON(bytes.NewBufferString(switcheo["arbitrum"]))
				ev, err := a.EventByID(common.HexToHash(UnLockEvent_Switcheo))
				if err != nil {
					log.Error("Poly Exact() can't decode LockEvent", "Chain", chain, "Hash", e.Hash)
				}
				ss, err := ev.Inputs.Unpack(hexutil.MustDecode(e.Data))
				d.ToAsset = strings.ToLower(fmt.Sprintf("%s", ss[0].(*big.Int).String()))
				res.ToAddress.Scan(strings.ToLower(fmt.Sprintf("0x%x", ss[1].([]uint8))))
			} else if len(e.Topics) == 1 {
				a, err := abi.JSON(bytes.NewBufferString(switcheo["eth"]))
				ev, err := a.EventByID(common.HexToHash(UnLockEvent_Switcheo))
				if err != nil {
					log.Error("Poly Exact() can't decode LockEvent", "Chain", chain, "Hash", e.Hash)
				}
				ss, err := ev.Inputs.Unpack(hexutil.MustDecode(e.Data))
				res.Token = strings.ToLower(ss[0].(common.Address).Hex())
				res.FromAddress.Scan(strings.ToLower(ss[1].(common.Address).Hex()))
				res.ToChainId = switchChainId(ss[2])
				d.ToAsset = strings.ToLower(fmt.Sprintf("0x%x", ss[3].([]uint8)))
				res.ToAddress.Scan(fmt.Sprintf("0x%x", ss[3].([]uint8)))
			}
			fromChainId := new(big.Int).Set(utils.GetChainId(chain))
			res.FromChainId = (*model.BigInt)(fromChainId)
			retLock = append(retLock, res)
		}
	}

	//处理跨入
	ret := make(model.Results, 0)
	for _, r := range retUnlock {
		for _, t := range retVerify {
			if t.Hash == r.Hash && r.ActionId < t.ActionId {
				r.FromChainId = t.FromChainId
				r.MatchTag = t.MatchTag
				r.ToChainId = t.ToChainId
				var a, b Detail
				_ = json.Unmarshal(r.Detail, &a)
				_ = json.Unmarshal(t.Detail, &b)
				a.CrossChainTxHash = b.CrossChainTxHash
				a.ToContract = b.ToContract
				r.Detail, _ = json.Marshal(a)
				break
			}
		}
		if r.FromChainId.String() == utils.GetChainId("switcheo").String() {
			r.Project = "SwitcheoNetwork"
		}
		ret = append(ret, r)
	}

	//处理跨出
	for _, r := range retLock {
		for _, t := range retCross {
			if r.Hash == t.Hash && r.ActionId > t.ActionId && r.ToChainId.String() == t.ToChainId.String() {
				r.MatchTag = t.MatchTag
				var a, b Detail
				_ = json.Unmarshal(r.Detail, &a)
				_ = json.Unmarshal(t.Detail, &b)
				a.ToContract = b.ToContract
				r.Detail, _ = json.Marshal(a)
				break
			}
		}
		if r.ToChainId.String() == utils.GetChainId("switcheo").String() {
			r.Project = "SwitcheoNetwork"
		}
		ret = append(ret, r)
	}
	return ret
}

func switchChainId(a interface{}) *model.BigInt {
	var oriId int
	switch a.(type) {
	case uint64:
		oriId = int(a.(uint64))
	case string:
		ori, _ := new(big.Int).SetString(a.(string), 16)
		oriId = int(ori.Int64())
	case *big.Int:
		oriId = int(a.(*big.Int).Int64())
	default:
		return nil
	}
	switch oriId {
	case 1:
		return (*model.BigInt)(utils.GetChainId("btc"))
	case 2:
		return (*model.BigInt)(utils.GetChainId("eth"))
	case 3:
		return (*model.BigInt)(utils.GetChainId("ontology"))
	case 4:
		return (*model.BigInt)(utils.GetChainId("neo"))
	case 5:
		return (*model.BigInt)(utils.GetChainId("switcheo"))
	case 6:
		return (*model.BigInt)(utils.GetChainId("bsc"))
	case 7:
		return (*model.BigInt)(utils.GetChainId("heco"))
	case 8:
		return (*model.BigInt)(utils.GetChainId("palette"))
	case 18:
		return (*model.BigInt)(utils.GetChainId("zilliqa"))
	case 10:
		return (*model.BigInt)(utils.GetChainId("curve"))
	case 12:
		return (*model.BigInt)(utils.GetChainId("okx"))
	case 14:
		return (*model.BigInt)(utils.GetChainId("neo"))
	case 15:
		return (*model.BigInt)(utils.GetChainId("heimdall"))
	case 17:
		return (*model.BigInt)(utils.GetChainId("polygon"))
	case 19:
		return (*model.BigInt)(utils.GetChainId("arbitrum"))
	case 20:
		return (*model.BigInt)(utils.GetChainId("gnosis"))
	case 21:
		return (*model.BigInt)(utils.GetChainId("avalanche"))
	case 22:
		return (*model.BigInt)(utils.GetChainId("fantom"))
	case 23:
		return (*model.BigInt)(utils.GetChainId("optimism"))
	case 24:
		return (*model.BigInt)(utils.GetChainId("metis"))
	case 25:
		return (*model.BigInt)(utils.GetChainId("boba"))
	case 26:
		return (*model.BigInt)(utils.GetChainId("oasis"))
	case 27:
		return (*model.BigInt)(utils.GetChainId("harmony"))
	case 28:
		return (*model.BigInt)(utils.GetChainId("hsc"))
	case 30:
		return (*model.BigInt)(utils.GetChainId("kcc"))
	case 32:
		return (*model.BigInt)(utils.GetChainId("kava"))
	case 35:
		return (*model.BigInt)(utils.GetChainId("cube"))
	case 36:
		return (*model.BigInt)(utils.GetChainId("celo"))
	case 40:
		return (*model.BigInt)(utils.GetChainId("astar"))
	default:
		return nil
	}
}
