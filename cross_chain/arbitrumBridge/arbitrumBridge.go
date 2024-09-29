package arbitrumBridge

import (
	"app/model"
	"app/utils"
	"math/big"
	"sort"
	"strings"
)

var _ model.EventCollector = &ArbiBridge{}

type ArbiBridge struct {
}

func NewArbiBridge() *ArbiBridge {
	return &ArbiBridge{}
}

func (a *ArbiBridge) Name() string {
	return "ArbitrumBridge"
}
func (a *ArbiBridge) Contracts(chain string) map[string]string {
	return make(map[string]string)
}

func (a *ArbiBridge) Topics0(chain string) []string {
	return []string{DepositInitiated, TxToL2, WithdrawalInitiated,
		OutBoxTransactionExecuted, WithdrawalFinalized, DepositFinalized}
}

func (a *ArbiBridge) SrcTopics0() []string {
	return []string{
		WithdrawalInitiated, DepositInitiated,
	}
}

var matchChain = map[string]string{
	"eth": "arbitrum", "arbitrum": "eth",
}

func (a *ArbiBridge) Extract(chain string, events model.Events) model.Results {
	ret := make(model.Results, 0)
	var seqL2Address = make(map[string]string)
	var outBoxHashSeq = make(map[string]idSeqs) //hash -> action_id -> seqNum
	for _, e := range events {
		if len(e.Topics) < 3 && e.Topics[0] != OutBoxTransactionExecuted {
			continue
		}
		var res *model.Result
		switch e.Topics[0] {
		//下面是从eth -> arbi
		case DepositInitiated:
			res = model.ScanBaseInfo(chain, a.Name(), e)
			res.Direction = model.OutDirection
			l1Token := "0x" + e.Data[26:66]
			res.Token = l1Token
			res.FromAddress.Scan("0x" + e.Topics[1][26:])
			res.ToAddress.Scan("0x" + e.Topics[2][26:])
			seq, _ := new(big.Int).SetString(e.Topics[3][2:], 16) //方便之后唯一对应
			amount, _ := new(big.Int).SetString(e.Data[66:130], 16)
			res.Amount = (*model.BigInt)(amount)
			res.MatchTag = amount.String() + l1Token + "," + seq.String() + e.Hash
			ret = append(ret, res)
		case TxToL2:
			seq, _ := new(big.Int).SetString(e.Topics[3][2:], 16) //方便之后唯一对应
			seqL2Address[seq.String()+e.Hash] = "0x" + e.Topics[2][26:]
		case DepositFinalized:
			res = model.ScanBaseInfo(chain, a.Name(), e)
			res.Direction = model.InDirection
			l1Token := "0x" + e.Topics[1][26:]
			res.FromAddress.Scan("0x" + e.Topics[2][26:])
			res.ToAddress.Scan("0x" + e.Topics[3][26:])
			amount, _ := new(big.Int).SetString(e.Data[2:66], 16)
			res.Amount = (*model.BigInt)(amount)
			res.MatchTag = amount.String() + l1Token + e.Address
			ret = append(ret, res)

			//下面是从arbi -> eth
		case WithdrawalInitiated:
			res = model.ScanBaseInfo(chain, a.Name(), e)
			res.Direction = model.OutDirection
			res.FromAddress.Scan("0x" + e.Topics[1][26:])
			res.ToAddress.Scan("0x" + e.Topics[2][26:])
			seq, _ := new(big.Int).SetString(e.Topics[3][2:], 16)
			l1Token := "0x" + e.Data[26:66]
			exitNum, _ := new(big.Int).SetString(e.Data[2+64:2+128], 16)
			amount, _ := new(big.Int).SetString(e.Data[2+128:], 16)
			res.Amount = (*model.BigInt)(amount)
			l2Sender := e.Address
			res.MatchTag = seq.String() + l2Sender + exitNum.String() + l1Token
			ret = append(ret, res)
		case WithdrawalFinalized:
			res = model.ScanBaseInfo(chain, a.Name(), e)
			res.Direction = model.InDirection
			res.FromAddress.Scan("0x" + e.Topics[1][26:])
			res.ToAddress.Scan("0x" + e.Topics[2][26:])
			exitNum, _ := new(big.Int).SetString(e.Topics[3][2:], 16)
			targetToken := "0x" + e.Data[26:66]
			amount, _ := new(big.Int).SetString(e.Data[66:130], 16)
			res.Amount = (*model.BigInt)(amount)
			res.MatchTag = exitNum.String() + targetToken
			ret = append(ret, res)
		case OutBoxTransactionExecuted:
			l2Sender := "0x" + e.Topics[2][26:]
			seq, _ := new(big.Int).SetString(e.Data[2:], 16)
			outBoxHashSeq[e.Hash] = append(outBoxHashSeq[e.Hash], &idSeq{
				e.Id, seq.String() + l2Sender,
			})
		}
	}

	for _, r := range ret {
		if r.Direction == model.InDirection {
			r.ToChainId = (*model.BigInt)(utils.GetChainId(r.Chain))
			if v, ok := OtherBridgeContractOnETH[r.Contract]; ok {
				r.FromChainId = (*model.BigInt)(utils.GetChainId(v))
			} else {
				r.FromChainId = (*model.BigInt)(utils.GetChainId(matchChain[r.Chain]))
			}

			if r.Chain == "eth" { //withdrawFinalized
				sort.Sort(outBoxHashSeq[r.Hash])
				for i, v := range outBoxHashSeq[r.Hash] {
					if v.actionId < r.ActionId {
						r.MatchTag = v.seqTag + r.MatchTag
						//有可能同一笔tx里面包含了多个log,需要删掉已经加入的log
						var newlist = idSeqs{}
						newlist = append(newlist, outBoxHashSeq[r.Hash][:i]...)
						newlist = append(newlist, outBoxHashSeq[r.Hash][i+1:]...)
						outBoxHashSeq[r.Hash] = newlist
						break
					}
				}
			}
		} else if r.Direction == model.OutDirection {
			r.FromChainId = (*model.BigInt)(utils.GetChainId(r.Chain))
			if v, ok := OtherBridgeContractOnETH[r.Contract]; ok {
				r.ToChainId = (*model.BigInt)(utils.GetChainId(v))
			} else {
				r.ToChainId = (*model.BigInt)(utils.GetChainId(matchChain[r.Chain]))
			}

			if r.Chain == "eth" { //depositInitiated
				ss := strings.Split(r.MatchTag, ",")
				if len(ss) > 1 {
					r.MatchTag = ss[0] + seqL2Address[ss[1]]
				}
			}
		}
	}
	return ret
}

type idSeq struct {
	actionId uint64
	seqTag   string
}

type idSeqs []*idSeq

func (a idSeqs) Len() int {
	return len(a)
}
func (t idSeqs) Less(a, b int) bool {
	return t[a].actionId < t[b].actionId
}

func (t idSeqs) Swap(a, b int) {
	t[a], t[b] = t[b], t[a]
}
