package avaxBridge

import (
	"app/model"
	"app/utils"
	"bytes"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
	"math/big"
	"strings"
)

var _ model.EventCollector = &AvaxBridgeEvent{}
var _ model.TransferCollector = &AvaxBridgeTransfer{}

type AvaxBridgeEvent struct{}
type AvaxBridgeTransfer struct {
	transferAddresses map[string][]string
}

func NewAvaxEventCollector() *AvaxBridgeEvent {
	return &AvaxBridgeEvent{}
}
func NewAvaxTransferCollector() *AvaxBridgeTransfer {
	return &AvaxBridgeTransfer{
		transferAddresses: map[string][]string{
			"eth": {AvaxBridgeContracts["eth"]},
		},
	}
}

func (a *AvaxBridgeEvent) Name() string {
	return "AvalancheBridge"
}
func (a *AvaxBridgeTransfer) Name() string {
	return "AvalancheBridge"
}

func (a *AvaxBridgeEvent) Contracts(chain string) map[string]string {
	return make(map[string]string)
}

func (a *AvaxBridgeEvent) Topics0(chain string) []string {
	if chain == "avalanche" {
		return []string{Mint}
	} else {
		return []string{}
	}
}

func (a *AvaxBridgeEvent) SrcTopics0() []string {
	return []string{}
}

func (a *AvaxBridgeEvent) Extract(chain string, events model.Events) model.Results {
	if _, ok := AvaxBridgeContracts[chain]; !ok {
		return nil
	}
	ret := make(model.Results, 0)
	for _, e := range events {
		var res = &model.Result{}
		switch e.Topics[0] {
		case Mint:
			res = model.ScanBaseInfo(chain, a.Name(), e)
			res.Direction = model.InDirection
			res.ToChainId = (*model.BigInt)(utils.GetChainId(chain))
			res.FromChainId = (*model.BigInt)(utils.GetChainId(matchChain[chain]))
			res.Token = e.Address
			a, err := abi.JSON(bytes.NewBufferString(avaxAbiStr))
			ev, err := a.EventByID(common.HexToHash(Mint))
			if err != nil {
				log.Error("OptiGateway Exact() can't decode MsgPassed", "Chain", chain, "Hash", e.Hash)
			}
			ss, err := ev.Inputs.Unpack(hexutil.MustDecode(e.Data))
			res.ToAddress.Scan(strings.ToLower(ss[0].(common.Address).String()))
			res.FromAddress = res.ToAddress
			res.Amount = (*model.BigInt)(ss[1].(*big.Int))
			res.MatchTag = fmt.Sprintf("0x%x", ss[4].([32]uint8))
		}
		ret = append(ret, res)
	}
	return ret
}

func (a *AvaxBridgeTransfer) Addresses(chain string) []string {
	return a.transferAddresses[chain]
}

// 目前只处理在eth上作为to地址的事件，其他要处理的时候需要补充
func (a *AvaxBridgeTransfer) Extract(chain string, msg model.ERC20Transfers) model.Results {
	var addr = make(map[string]struct{})
	for _, add := range a.Addresses(chain) {
		addr[add] = struct{}{}
	}
	var lastBlock, lastIndex, lastActionId uint64
	var ret model.Results
	for _, m := range msg {
		if lastBlock != m.Number {
			lastBlock = m.Number
			lastIndex = m.Index
			lastActionId = m.ActionId
		} else if lastIndex != m.Index {
			lastIndex = m.Index
			lastActionId = m.ActionId
		} else if lastActionId >= m.ActionId {
			m.ActionId = lastActionId + 1
			lastActionId = m.ActionId
		}
		var res = model.ScanBaseInfo(chain, a.Name(), m)
		if _, ok := addr[m.To]; ok {
			res.Direction = model.OutDirection
			res.FromAddress.Scan(m.From)
			res.ToAddress = res.FromAddress
			res.FromChainId = (*model.BigInt)(utils.GetChainId(chain))
			res.ToChainId = (*model.BigInt)(utils.GetChainId(matchChain[chain]))
		} else if _, ok := addr[m.From]; ok {
			continue //暂时只处理eth作为转出的
			/*res.Direction = model.InDirection
			res.ToAddress.Scan(m.To)
			res.FromAddress = res.FromAddress
			res.FromChainId = (*model.BigInt)(utils.GetChainId(chain))
			res.ToChainId = (*model.BigInt)(utils.GetChainId(matchChain[chain]))*/
		} else {
			continue
		}
		res.Amount = (*model.BigInt)(m.Value)
		res.Token = m.ContractAddress
		res.MatchTag = m.Hash
		ret = append(ret, res)
	}
	return ret
}
