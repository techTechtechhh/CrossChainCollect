package optimismGateway

import (
	"app/model"
	"app/utils"
	"bytes"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"math/big"
	"strings"
)

var _ model.EventCollector = &OptiGateway{}

type OptiGateway struct{}

const (
	FuncRelayMsg  = "relayMessage"
	FuncRelayMsg0 = "relayMessage0"
)

func NewOptiCollector() *OptiGateway {
	return &OptiGateway{}
}

func (a *OptiGateway) Name() string {
	return "OptimismGateway"
}

func (a *OptiGateway) Contracts(chain string) map[string]string {
	/*if _, ok := OptiContracts[chain]; !ok {
		return nil
	}
	return OptiContracts[chain]*/
	return make(map[string]string)
}

func (a *OptiGateway) Topics0(chain string) []string {
	return []string{RelayedMessage, DepositFinalized, MessagePassed, SentMessage, WithdrawalInitiated,
		SentMessageExtension1, ETHDepositInitiated, ERC20DepositInitiated, ERC20WithdrawalFinalized, ETHWithdrawalFinalized}
}

func (a *OptiGateway) SrcTopics0() []string {
	return []string{SentMessage, WithdrawalInitiated, ETHDepositInitiated, ERC20DepositInitiated}
}

type msgIndex struct {
	Msg   string
	Index uint64
	*param
}
type param struct {
	Nonce  *big.Int
	Sender common.Address
	Target common.Address
	Value  *big.Int
	Gas    *big.Int
	Data   []byte
}

func (a *OptiGateway) Extract(chain string, events model.Events) model.Results {
	ret := make(model.Results, 0)
	realRet := make(model.Results, 0)
	var HashMsg = make(map[string][]*msgIndex)
	var opNonceMap = make(map[string]*big.Int)
	var ethSenderMap = make(map[string]*big.Int)
	var inMatchTag = make(map[string]string)
	for _, e := range events {
		var res *model.Result
		switch e.Topics[0] {
		//optimism -> eth
		case WithdrawalInitiated:
			res = model.ScanBaseInfo(chain, a.Name(), e)
			res.Direction = model.OutDirection
			res.Token = "0x" + e.Topics[2][26:]
			res.FromAddress.Scan("0x" + e.Topics[3][26:])
			res.ToAddress.Scan("0x" + e.Data[26:66])
			res.Amount = new(model.BigInt).SetString(e.Data[66:130], 16)
			ret = append(ret, res)
		case MessagePassed:
			a, err := abi.JSON(bytes.NewBufferString(L2ToL1MessagePasser["optimism"]))
			ev, err := a.EventByID(common.HexToHash(MessagePassed))
			if err != nil {
				log.Error("OptiGateway Exact() can't decode MsgPassed", "Chain", chain, "Hash", e.Hash)
			}
			ss, err := ev.Inputs.Unpack(hexutil.MustDecode(e.Data))
			_nonce, _ := new(big.Int).SetString(e.Topics[1][2:], 16)
			_value := ss[0].(*big.Int)
			opNonceMap[_nonce.String()] = _value
			/*_sender := common.HexToAddress("0x" + e.Topics[2][26:])
			_target := common.HexToAddress("0x" + e.Topics[3][26:])*/
			/*_gasLimit := ss[1].(*big.Int)
			_data := common.Hex2Bytes(fmt.Sprintf("%x", ss[2].([]uint8)))*/
		case SentMessage:
			a, err := abi.JSON(bytes.NewBufferString(L2CrossDomainMessenger["optimism"]))
			ev, err := a.EventByID(common.HexToHash(SentMessage))
			if err != nil {
				log.Error("OptiGateway Exact() can't decode SentMsg", "Chain", chain, "Hash", e.Hash)
			}
			ss, err := ev.Inputs.Unpack(hexutil.MustDecode(e.Data))
			_target := common.HexToAddress("0x" + e.Topics[1][26:])
			_sender := ss[0].(common.Address)
			_data := ss[1].([]uint8)
			_nonce := ss[2].(*big.Int)
			_gasLimit := ss[3].(*big.Int)
			HashMsg[e.Hash] = append(HashMsg[e.Hash], &msgIndex{
				fmt.Sprintf("%x", _data), e.Index,
				&param{
					Nonce:  _nonce,
					Sender: _sender,
					Target: _target,
					Gas:    _gasLimit,
					Data:   _data,
				},
			})

		case RelayedMessage:
			if len(e.Topics) >= 2 {
				inMatchTag[e.Hash] = e.Topics[1]
			} else if len(e.Data) > 2 {
				inMatchTag[e.Hash] = e.Data
			}
		case ERC20WithdrawalFinalized:
			res = model.ScanBaseInfo(chain, a.Name(), e)
			res.Direction = model.InDirection
			res.Token = "0x" + e.Topics[1][26:]
			res.FromAddress.Scan("0x" + e.Topics[3][26:])
			res.ToAddress.Scan("0x" + e.Data[26:66])
			res.Amount = new(model.BigInt).SetString(e.Data[66:130], 16)
			ret = append(ret, res)
		case ETHWithdrawalFinalized:
			res = model.ScanBaseInfo(chain, a.Name(), e)
			res.Direction = model.InDirection
			res.Token = model.NativeToken
			res.FromAddress.Scan("0x" + e.Topics[1][26:])
			res.ToAddress.Scan("0x" + e.Topics[2][26:])
			res.Amount = new(model.BigInt).SetString(e.Data[2:66], 16)
			ret = append(ret, res)

		//eth -> Optimism
		case SentMessageExtension1:
			sender := "0x" + e.Topics[1][26:]
			value, _ := new(big.Int).SetString(e.Data[2:], 16)
			ethSenderMap[e.Hash+sender] = value
		case ETHDepositInitiated:
			res = model.ScanBaseInfo(chain, a.Name(), e)
			res.Direction = model.OutDirection
			res.FromAddress.Scan("0x" + e.Topics[1][26:])
			res.ToAddress.Scan("0x" + e.Topics[2][26:])
			res.Amount = new(model.BigInt).SetString(e.Data[2:66], 16)
			res.Token = model.NativeToken
			ret = append(ret, res)
		case DepositFinalized:
			res = model.ScanBaseInfo(chain, a.Name(), e)
			res.Direction = model.InDirection
			res.Token = "0x" + e.Topics[2][2:]
			res.FromAddress.Scan("0x" + e.Topics[3][26:])
			res.ToAddress.Scan("0x" + e.Data[26:66])
			res.Amount = new(model.BigInt).SetString(e.Data[66:130], 16)
			ret = append(ret, res)
		case ERC20DepositInitiated:
			res = model.ScanBaseInfo(chain, a.Name(), e)
			res.Direction = model.OutDirection
			res.Token = "0x" + e.Topics[1][26:]
			res.FromAddress.Scan("0x" + e.Topics[3][26:])
			res.ToAddress.Scan("0x" + e.Data[26:66])
			res.Amount = new(model.BigInt).SetString(e.Data[66:130], 16)
			ret = append(ret, res)
		}
	}

	for _, r := range ret {
		if r.Direction == model.InDirection {
			r.ToChainId = (*model.BigInt)(utils.GetChainId(r.Chain))
			r.MatchTag = inMatchTag[r.Hash]
			realRet = append(realRet, r)
			continue
		}

		//处理转出的交易
		r.FromChainId = (*model.BigInt)(utils.GetChainId(r.Chain))
		for _, v := range HashMsg[r.Hash] {
			if strings.Contains(v.Msg, r.ToAddress.String[2:]) {
				if r.Chain == "optimism" {
					v.Value = opNonceMap[v.Nonce.String()]
				} else if r.Chain == "eth" {
					v.Value = ethSenderMap[r.Hash+strings.ToLower(v.Sender.String())]
				}

				var tag []byte
				if v.Value != nil {
					contractABI, err := abi.JSON(strings.NewReader(L2CrossDomainMessenger["optimism"]))
					if err != nil {
						fmt.Println("解析 ABI 错误:", err)
					}
					tag, err = contractABI.Pack(FuncRelayMsg, v.Nonce, v.Sender, v.Target, v.Value, v.Gas, v.Data)
					if err != nil {
						fmt.Println("生成函数调用数据错误:", err)
					}
				} else {
					contractABI, err := abi.JSON(strings.NewReader(L1CrossDomainMessenger["eth"]))
					if err != nil {
						fmt.Println("解析 ABI 错误:", err)
					}
					tag, err = contractABI.Pack(FuncRelayMsg, v.Target, v.Sender, v.Data, v.Nonce)
					if err != nil {
						fmt.Println("生成函数调用数据错误:", err)
					}
				}
				r.MatchTag = crypto.Keccak256Hash(tag).String()
				realRet = append(realRet, r)
				break
			}
		}
	}
	return realRet
}
