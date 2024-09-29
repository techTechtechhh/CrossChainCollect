package geth

import (
	"app/model"
	"app/utils"
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"golang.org/x/time/rate"
)

type GethProvider struct {
	chain     string
	client    *ethclient.Client
	callCache sync.Map
	limiter   *rate.Limiter
}

func NewGethProvider(chain, addr string) *GethProvider {
	c, err := ethclient.Dial(addr)
	if err != nil {
		panic(err)
	}
	return &GethProvider{
		chain:   chain,
		client:  c,
		limiter: rate.NewLimiter(20, 1),
	}
}

func (p *GethProvider) Call(from, to, input string, value *big.Int, number *big.Int) ([]byte, error) {
	callKey := genMsgCallKey(from, to, input, value, number)
	if val, ok := p.callCache.Load(callKey); ok {
		ret, _ := val.([]byte)
		return ret, nil
	}
	var toAddr *common.Address
	if to != "" {
		tmp := common.HexToAddress(to)
		toAddr = &tmp
	}
	msg := ethereum.CallMsg{
		From:  common.HexToAddress(from),
		To:    toAddr,
		Value: value,
		Data:  common.FromHex(input),
	}
	p.limiter.Wait(context.Background())
	ret, err := p.client.CallContract(context.Background(), msg, number)
	if err == nil {
		p.callCache.Store(callKey, ret)
	}
	return ret, err
}

func (p *GethProvider) ContinueCall(from, to, input string, value *big.Int, number *big.Int) ([]byte, error) {
	var err error
	var ret []byte
	for {
		ret, err = p.Call(from, to, input, value, number)
		if !utils.IsNetError(err) {
			break
		}
		log.Warn("msg call failed due to net error, retrying", "chain", p.chain)
		time.Sleep(time.Second * 5)
		break
	}
	return ret, err
}

func (p *GethProvider) LatestNumber() (uint64, error) {
	p.limiter.Wait(context.Background())
	return p.client.BlockNumber(context.Background())
}

func (p *GethProvider) GetLogs(topics0 []string, from, to uint64) (model.Events, error) {
	ret := make(model.Events, 0)
	blockToTs := make(map[uint64]time.Time, 0)

	qry := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(from),
		ToBlock:   new(big.Int).SetUint64(to),
		Topics:    make([][]common.Hash, 0),
	}
	topic0 := make([]common.Hash, 0)
	for _, t := range topics0 {
		if len(t) != 0 {
			topic0 = append(topic0, common.HexToHash(t))
		}
	}
	if len(topic0) != 0 {
		qry.Topics = append(qry.Topics, topic0)
	}
	p.limiter.Wait(context.Background())
	rawLogs, err := p.client.FilterLogs(context.Background(), qry)
	if err != nil && strings.Contains(strings.ToLower(err.Error()), "response size should not greater than") {
		return nil, utils.ErrGethRespTooLarge
	}
	if err != nil {
		return nil, err
	}

	for _, rawLog := range rawLogs {
		if rawLog.Removed {
			continue
		}
		topics := make([]string, 0)
		for _, t := range rawLog.Topics {
			topics = append(topics, hexutil.Encode(t[:]))
		}
		if _, ok := blockToTs[rawLog.BlockNumber]; !ok {
			block, err := p.client.BlockByHash(context.Background(), rawLog.BlockHash)
			if err != nil {
				blockToTs[rawLog.BlockNumber] = time.Unix(time.Now().Unix(), 0)
			} else {
				blockToTs[rawLog.BlockNumber] = time.Unix(int64(block.Time()), 0)
			}
		}
		ret = append(ret, &model.Event{
			Number:  rawLog.BlockNumber,
			Ts:      blockToTs[rawLog.BlockNumber],
			Index:   uint64(rawLog.TxIndex),
			Hash:    hexutil.Encode(rawLog.TxHash[:]),
			Id:      uint64(rawLog.Index),
			Address: strings.ToLower(rawLog.Address.Hex()),
			Topics:  topics,
			Data:    hexutil.Encode(rawLog.Data),
		})
	}
	return ret, nil
}

func genMsgCallKey(from, to, input string, value *big.Int, number *big.Int) string {
	return from + to + input + value.String() + number.String()
}

func (p *GethProvider) GetOriginLogs(topics0 string, is1or2 int, topics1or2 string, from, to uint64) (model.Events, error) {
	ret := make(model.Events, 0)
	qry := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(from),
		ToBlock:   new(big.Int).SetUint64(to),
		Topics:    make([][]common.Hash, 0),
	}
	topic0 := make([]common.Hash, 0)
	topics0_1 := make([]common.Hash, 0)
	topics0_2 := make([]common.Hash, 0)
	topic0 = append(topic0, common.HexToHash(topics0))
	if is1or2 == 1 {
		topics0_1 = append(topics0_1, common.HexToHash(topics1or2))
	} else {
		topics0_2 = append(topics0_2, common.HexToHash(topics1or2))
	}
	qry.Topics = append(qry.Topics, topic0, topics0_1, topics0_2)

	p.limiter.Wait(context.Background())
	rawLogs, err := p.client.FilterLogs(context.Background(), qry)
	if err != nil {
		return nil, err
	}

	for _, rawLog := range rawLogs {
		if rawLog.Removed {
			continue
		}
		topics := make([]string, 0)
		for _, t := range rawLog.Topics {
			topics = append(topics, hexutil.Encode(t[:]))
		}
		ret = append(ret, &model.Event{
			Number:  rawLog.BlockNumber,
			Index:   uint64(rawLog.TxIndex),
			Hash:    hexutil.Encode(rawLog.TxHash[:]),
			Id:      uint64(rawLog.Index),
			Address: strings.ToLower(rawLog.Address.Hex()),
			Topics:  topics,
			Data:    hexutil.Encode(rawLog.Data),
		})
	}
	return ret, nil
}

func (p *GethProvider) GetTxInfoByTxHash(txHash string) (*types.Transaction, error) {
	tx, _, err := p.client.TransactionByHash(context.Background(), common.HexToHash(txHash))
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (p *GethProvider) GetTxSender(txHash string, blockNumber uint64, index uint) (string, error) {
	tx, err := p.GetTxInfoByTxHash(txHash)
	if err != nil {
		err = fmt.Errorf("Geth get txInfoByHash failed when getting tx sender: %s", err)
		return "", err
	}
	block, err := p.client.BlockByNumber(context.Background(), new(big.Int).SetUint64(blockNumber))
	if err != nil {
		err = fmt.Errorf("Geth get blockHashByNumber failed when getting tx sender: %s", err)
		return "", err
	}
	sender, err := p.client.TransactionSender(context.Background(), tx, block.Hash(), index)
	if err != nil {
		err = fmt.Errorf("Geth get TransactionSender failed when getting tx sender: %s", err)
		return "", err
	}
	return sender.String(), err
}
