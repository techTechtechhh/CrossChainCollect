package wormhole

import (
	"app/model"
	"app/svc"
	"app/utils"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
)

type WormHole struct {
	svc *svc.ServiceContext
}

var _ model.MsgCollector = &WormHole{}

func NewWormHoleCollector(svc *svc.ServiceContext) *WormHole {
	return &WormHole{
		svc: svc,
	}
}

func (w *WormHole) Name() string {
	return "WormHole"
}

func (w *WormHole) Contracts(chain string) []string {
	if _, ok := contracts[chain]; !ok {
		return nil
	}
	addrs := make([]string, 0)
	for addr := range contracts[chain] {
		addrs = append(addrs, strings.ToLower(addr))
	}
	return addrs
}

func (w *WormHole) Selectors(chain string) []string {
	return []string{TransferTokens, WrapAndTransferETH, TransferTokensWithPayload, WrapAndTransferETHWithPayload,
		CompleteTransfer, CompleteTransferAndUnwrapETH, CompleteTransferWithPayload, CompleteTransferAndUnwrapETHWithPayload}
}

func (w *WormHole) Extract(chain string, msgs []*model.Call) model.Results {
	if _, ok := contracts[chain]; !ok {
		return nil
	}
	ret := make(model.Results, 0)
	for _, msg := range msgs {
		if _, ok := contracts[chain][msg.To]; !ok {
			continue
		}
		if len(msg.Input) <= 10 {
			continue
		}
		sig, rawParam := msg.Input[:10], msg.Input[10:]
		params, err := utils.DecodeInput(wormAbi, sig, rawParam)
		if err != nil {
			log.Error("decode wormwhole failed", "chain", chain, "hash", msg.Hash, "err", err)
			continue
		}
		res := &model.Result{
			Chain:    chain,
			Number:   msg.Number,
			Ts:       msg.Ts.UTC(),
			Index:    msg.Index,
			Hash:     msg.Hash,
			ActionId: msg.Id,
			Project:  w.Name(),
			Contract: msg.To,
		}
		switch sig {
		case TransferTokens, TransferTokensWithPayload:
			if len(params) < 6 {
				log.Error("decode wormwhole failed", "chain", chain, "hash", msg.Hash)
				continue
			}
			res.Direction = model.OutDirection
			res.FromChainId = (*model.BigInt)(utils.GetChainId(chain))
			res.FromAddress.Scan(msg.From)
			toChain, ok := params[2].(uint16)
			if !ok {
				log.Error("decode wormwhole failed", "chain", chain, "hash", msg.Hash)
			}
			res.ToChainId = (*model.BigInt)(new(big.Int).SetUint64(ConvertChainId(uint64(toChain))))
			to, ok := params[3].([32]byte)
			if !ok {
				log.Error("decode wormwhole failed", "chain", chain, "hash", msg.Hash)
				continue
			}
			res.ToAddress.Scan(truncateAddress(hexutil.Encode(to[:])))
			token, ok := params[0].(common.Address)
			if !ok {
				log.Error("decode wormwhole failed", "chain", chain, "hash", msg.Hash)
				continue
			}
			res.Token = strings.ToLower(token.String())
			if _, ok := WormholeTokenDecimals[chain]; !ok {
				WormholeTokenDecimals[chain] = map[string]*big.Int{}
			}
			if _, ok := WormholeTokenDecimals[chain][res.Token]; !ok {
				decimals, err := w.GetDecimal(chain, res.Token)
				if err != nil {
					time.Sleep(time.Second)
					log.Error("wormwhole get decimals failed", "chain", chain, "hash", msg.Hash, "err", err)
					continue
				}
				WormholeTokenDecimals[chain][res.Token] = decimals
			}
			amount, ok := params[1].(*big.Int)
			if !ok {
				log.Error("decode wormwhole failed", "chain", chain, "hash", msg.Hash)
				continue
			}
			// 2023-5-9
			// bug fix: https://upgradehub.xyz/diffs/etherscan/0x3ee18b2214aff97000d974cf647e7c347e8fa585?selected=3
			// amount = deNormalizeAmount(amount, uint8(WormholeTokenDecimals[chain][res.Token].Uint64()))
			res.Amount = (*model.BigInt)(amount)
			var nonce uint32
			if sig == TransferTokens {
				nonce, ok = params[5].(uint32)
			} else {
				nonce, ok = params[4].(uint32)
			}
			if !ok {
				log.Error("decode wormwhole failed", "chain", chain, "hash", msg.Hash)
				continue
			}
			res.MatchTag = strconv.FormatUint(uint64(nonce), 10)
			d := &OutDetail{Nonce: nonce}
			res.Detail, _ = json.Marshal(d)

		case WrapAndTransferETH, WrapAndTransferETHWithPayload:
			res.Direction = model.OutDirection
			res.FromChainId = (*model.BigInt)(utils.GetChainId(chain))
			res.FromAddress.Scan(msg.From)
			toChain, ok := params[0].(uint16)
			if !ok {
				log.Error("decode wormwhole failed", "chain", chain, "hash", msg.Hash)
			}
			res.ToChainId = (*model.BigInt)(new(big.Int).SetUint64(ConvertChainId(uint64(toChain))))
			to, ok := params[1].([32]byte)
			if !ok {
				log.Error("decode wormwhole failed", "chain", chain, "hash", msg.Hash)
				continue
			}
			res.ToAddress.Scan(truncateAddress(hexutil.Encode(to[:])))

			res.Token = model.NativeToken
			res.Amount = (*model.BigInt)(new(big.Int).Set(msg.Value))
			var nonce uint32
			if sig == WrapAndTransferETH {
				nonce, ok = params[3].(uint32)
			} else {
				nonce, ok = params[2].(uint32)
			}
			if !ok {
				log.Error("decode wormwhole failed", "chain", chain, "hash", msg.Hash)
				continue
			}
			res.MatchTag = strconv.FormatUint(uint64(nonce), 10)
			d := &OutDetail{Nonce: nonce}
			res.Detail, _ = json.Marshal(d)

		case CompleteTransfer, CompleteTransferAndUnwrapETH, CompleteTransferWithPayload, CompleteTransferAndUnwrapETHWithPayload:
			res.Direction = model.InDirection
			res.ToChainId = (*model.BigInt)(utils.GetChainId(chain))
			if len(params) == 0 {
				log.Error("decode wormwhole vm failed", "chain", chain, "hash", msg.Hash)
				continue
			}
			vm, ok := params[0].([]byte)
			if !ok {
				log.Error("decode wormwhole vm failed", "chain", chain, "hash", msg.Hash)
				continue
			}
			vaa := ParseVAA(vm)
			if vaa == nil {
				log.Error("decode wormwhole vaa failed", "chain", chain, "hash", msg.Hash)
				continue
			}
			transferPayload := ParseTokenTransferPayload(common.FromHex(vaa.Payload))
			if transferPayload == nil {
				log.Error("decode wormwhole vaa failed", "chain", chain, "hash", msg.Hash)
				continue
			}
			chainId, err := w.GetChainId(chain, msg.To)
			if err != nil {
				log.Error("wormwhole get chainid failed", "chain", chain, "hash", msg.Hash, "transferPayload ToChain", transferPayload.ToChain, "err", err)
				continue
			}
			if chainId != transferPayload.ToChain {
				log.Error("wormwhole invalid target chain", "chain", chain, "hash", msg.Hash)
				continue
			}
			if transferPayload.TokenChain == chainId {
				res.Token = truncateAddress(transferPayload.TokenAddress)
			} else {
				wrapped, err := w.GetWrappedAsset(chain, msg.To, transferPayload.TokenChain, common.FromHex(transferPayload.TokenAddress))
				if err != nil {
					log.Error("wormwhole get wrapped asset failed", "chain", chain, "hash", msg.Hash, "err", err)
					continue
				}
				res.Token = strings.ToLower(wrapped.String())
			}

			if _, ok := WormholeTokenDecimals[chain]; !ok {
				WormholeTokenDecimals[chain] = map[string]*big.Int{}
			}
			if _, ok := WormholeTokenDecimals[chain][res.Token]; !ok {
				time.Sleep(time.Second)
				decimals, err := w.GetDecimal(chain, res.Token)
				if err != nil {
					log.Error("wormwhole get decimals failed", "chain", chain, "hash", msg.Hash, "err", err)
					continue
				}
				WormholeTokenDecimals[chain][res.Token] = decimals
			}
			// 2023-5-9
			// bug fix: https://upgradehub.xyz/diffs/etherscan/0x3ee18b2214aff97000d974cf647e7c347e8fa585?selected=3
			res.Amount = (*model.BigInt)(deNormalizeAmount(transferPayload.Amount, uint8(WormholeTokenDecimals[chain][res.Token].Uint64())))
			res.ToAddress.Scan(truncateAddress(transferPayload.To))
			res.FromAddress.Scan(truncateAddress(transferPayload.FromAddress))
			res.MatchTag = strconv.FormatUint(uint64(vaa.Nonce), 10)
			detail, _ := json.Marshal(vaa)
			res.Detail = detail
		default:
			continue
		}
		ret = append(ret, res)
	}
	return ret
}

func (w *WormHole) GetDecimal(chain, token string) (*big.Int, error) {
	p := w.svc.Providers.Get(chain)
	if p == nil {
		return nil, fmt.Errorf("providers does not support %v", chain)
	}
	raw, err := p.ContinueCall("", token, "0x313ce567", nil, nil)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return nil, fmt.Errorf("invalid decimals() return, len = 0")
	}
	return new(big.Int).SetBytes(raw), nil
}

func (w *WormHole) GetChainId(chain, bridge string) (uint16, error) {
	p := w.svc.Providers.Get(chain)
	if p == nil {
		return 0, fmt.Errorf("providers does not support %v", chain)
	}
	raw, err := p.ContinueCall("", bridge, "0x9a8a0592", nil, nil)
	if err != nil {
		return 0, err
	}
	if len(raw) < 32 {
		return 0, fmt.Errorf("invalid chain id return, len = 0")
	}
	return binary.BigEndian.Uint16(raw[30:32]), nil
}

func (w *WormHole) GetWrappedAsset(chain, bridge string, tokenChainId uint16, TokenAddress []byte) (common.Address, error) {
	p := w.svc.Providers.Get(chain)
	if p == nil {
		return common.Address{}, fmt.Errorf("providers does not support %v", chain)
	}
	p1 := make([]byte, 2)
	binary.BigEndian.PutUint16(p1, tokenChainId)
	p1 = common.LeftPadBytes(p1, 32)
	TokenAddress = common.LeftPadBytes(TokenAddress, 32)

	raw, err := p.ContinueCall("", bridge, "0x1ff1e286"+common.Bytes2Hex(p1)+common.Bytes2Hex(TokenAddress), nil, nil)
	if err != nil {
		return common.Address{}, err
	}
	if len(raw) == 0 {
		return common.Address{}, fmt.Errorf("invalid wrapped asset return, len = 0")
	}
	return common.BytesToAddress(raw), nil
}
