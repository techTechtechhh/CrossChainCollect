package synapse

import (
	"app/model"
	"app/svc"
	"app/utils"
	"encoding/json"
	"github.com/ethereum/go-ethereum/log"
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/crypto"
)

var _ model.EventCollector = &Synapse{}

type Synapse struct {
	svc *svc.ServiceContext
}

func NewSynapseCollector(svc *svc.ServiceContext) *Synapse {
	initSynapse()
	return &Synapse{svc}
}

func (a *Synapse) Name() string {
	return "Synapse"
}

func (a *Synapse) Contracts(chain string) map[string]string {
	return make(map[string]string)
	/*if _, ok := SynapseContracts[chain]; !ok {
		return nil
	}
	return SynapseContracts[chain]*/
}

func (a *Synapse) Topics0(chain string) []string {
	return []string{TokenDeposit, TokenDepositAndSwap, TokenMint,
		TokenMintAndSwap, TokenRedeem, TokenRedeemAndRemove, TokenRedeemAndSwap,
		TokenWithdraw, TokenWithdrawAndRemove}
}

func (a *Synapse) SrcTopics0() []string {
	return []string{
		TokenDeposit, TokenDepositAndSwap, TokenMint, TokenMintAndSwap,
	}
}

func (a *Synapse) Extract(chain string, events model.Events) model.Results {
	ret := make(model.Results, 0)
	var kappa string
	for _, e := range events {
		if len(e.Topics) == 1 {
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

		res.ToAddress.Scan("0x" + e.Topics[1][26:])
		// quick fix
		// TODO: track cross-in by function call
		res.FromAddress = res.ToAddress

		if e.Topics[0] == TokenDeposit || e.Topics[0] == TokenDepositAndSwap ||
			e.Topics[0] == TokenRedeem || e.Topics[0] == TokenRedeemAndRemove || e.Topics[0] == TokenRedeemAndSwap {
			if len(e.Topics) < 2 {
				continue
			}
			if len(e.Data) < 194 {
				log.Warn("Synapse abnormal tx: len(data)<194", "Hash", e.Hash, "Chain", chain)
				continue
			}
			fromChainId := new(big.Int).Set(utils.GetChainId(chain))
			res.FromChainId = (*model.BigInt)(fromChainId)
			res.Direction = model.OutDirection
			res.Token = "0x" + e.Data[2+64+24:2+128]
			toChainId, _ := new(big.Int).SetString(e.Data[2:2+64], 16)
			res.ToChainId = (*model.BigInt)(toChainId)
			amount, _ := new(big.Int).SetString(e.Data[2+128:2+192], 16)
			res.Amount = (*model.BigInt)(amount)
			var t = crypto.Keccak256Hash([]byte(res.Hash)).String()
			kappa = t
		}
		if e.Topics[0] == TokenMint || e.Topics[0] == TokenMintAndSwap ||
			e.Topics[0] == TokenWithdraw || e.Topics[0] == TokenWithdrawAndRemove {
			res.Direction = model.InDirection
			res.Token = "0x" + e.Data[2+24:2+64]
			amount, _ := new(big.Int).SetString(e.Data[2+64:2+128], 16)
			res.Amount = (*model.BigInt)(amount)
			toChainId := new(big.Int).Set(utils.GetChainId(chain))
			res.ToChainId = (*model.BigInt)(toChainId)
			kappa = e.Topics[len(e.Topics)-1]
		}
		d := &Detail{
			Kappa: kappa,
		}
		detail, err := json.Marshal(d)
		if err == nil {
			res.Detail = detail
		}
		res.MatchTag = kappa

		if res.Token == nUSD[res.Chain] {
			provider := a.svc.Providers.Get(res.Chain)
			swapEvents := getSynapseLogs(provider, []string{TokenSwap, RemoveLiquidity, AddLiquidity}, res.Hash)
			if swapEvents == nil {
				continue
			}
			sort.Sort(swapEvents)
			realToken := extractRealToken(res, swapEvents)
			if realToken != "" {
				res.Token = realToken
			}
		}

		ret = append(ret, res)
	}
	return ret
}
