package provider

import (
	"app/config"
	"app/model"
	"app/provider/chainbase"
	"app/provider/etherscan"
	"app/provider/geth"
	"fmt"
	"github.com/ethereum/go-ethereum/log"
	"math/big"
	"strings"
)

type Provider struct {
	chainName string
	geth      *geth.GethProvider
	scan      *etherscan.EtherscanProvider
	chainbase *chainbase.Provider
}

func (p *Provider) GetScan() *etherscan.EtherscanProvider {
	return p.scan
}

func (p *Provider) Call(from, to, input string, value *big.Int, number *big.Int) ([]byte, error) {
	return p.geth.Call(from, to, input, value, number)
}

func (p *Provider) ContinueCall(from, to, input string, value *big.Int, number *big.Int) ([]byte, error) {
	return p.geth.ContinueCall(from, to, input, value, number)
}

func (p *Provider) GetContractFirstInvocation(address string) (uint64, error) {
	val, err := p.scan.GetContractFirstInvocation(address)
	if err == nil {
		return val, nil
	}
	if p.chainbase == nil {
		return 0, err
	}
	return p.chainbase.GetContractFirstCreatedNumber(address)
}

func (p *Provider) LatestNumber() (uint64, error) {
	var val uint64
	var err error

	val, err = p.scan.LatestNumber()
	if err == nil {
		return val, err
	}
	log.Error("failed get latest number", "ERR", err)
	val, err = p.geth.LatestNumber()
	if err != nil {
		return 0, err
	}
	if val == 0 {
		return 0, fmt.Errorf("invalid latest block")
	}
	return val - 128, nil
}

func (p *Provider) GetLogs(topics0 []string, from, to uint64) (model.Events, error) {
	//return p.chainbase.GetLogs(topics0, from, to)
	return p.scan.GetLogs(topics0, from, to)
}

func (p *Provider) GetCalls(addresses, selectors []string, from, to uint64) ([]*model.Call, error) {
	// return p.scan.GetCalls(addresses, selectors, from, to)
	if p.chainbase == nil {
		return nil, nil
	}
	return p.chainbase.GetCalls(addresses, selectors, from, to)
}

func (p *Provider) GetERC20Transfer(address []string, from, to uint64) (model.ERC20Transfers, error) {
	return p.scan.GetERC20Transfer(address, from, to)
}

type Providers struct {
	providers map[string]*Provider
}

func NewProviders(cfg *config.Config) *Providers {
	providers := make(map[string]*Provider)
	for chainName, providerCfg := range cfg.ChainProviders {
		gethP := geth.NewGethProvider(chainName, providerCfg.Node)
		scanP := etherscan.NewEtherScanProvider(providerCfg.ScanUrl, providerCfg.ApiKeys, cfg.Proxy, cfg.EtherscanRateLimit)
		providers[chainName] = &Provider{
			chainName: chainName,
			geth:      gethP,
			scan:      scanP,
		}
		if providerCfg.ChainbaseTable != "" {
			providers[chainName].chainbase = chainbase.NewProvider(providerCfg.ChainbaseTable, cfg.ChainbaseApiKey, providerCfg.EnableTraceCall, cfg.Proxy)
		}
	}
	return &Providers{providers: providers}
}

func (p *Providers) Get(chain string) *Provider {
	if val, ok := p.providers[chain]; ok {
		return val
	}
	return nil
}

func (p *Providers) GetAll() map[string]*Provider {
	return p.providers
}

func (p *Providers) GetTransferInfo(chain string, topics0 string, is1or2 int, topics1or2 string, from, to uint64) ([]*model.TransferInfo, error) {
	ret := make([]*model.TransferInfo, 0)

	tfs, err := p.Get(chain).GetScan().GetOriginLogs(topics0, is1or2, topics1or2, etherscan.Option{
		Page:       1,
		PageSize:   10,
		StartBlock: int64(from),
		EndBlock:   int64(to),
	})
	if err != nil {
		return ret, err
	}
	for _, tf := range tfs {
		ret = append(ret, &model.TransferInfo{
			Token:       tf.Address,
			FromAddress: "0x" + tf.Topics[1][26:],
			ToAddress:   "0x" + tf.Topics[2][26:],
			Data:        tf.Data,
		})
	}
	return ret, err
}

func (p *Providers) GetTxInfo(chain string, txHash string) (*model.TxInfo, error) {
	txinfo, err := p.Get(chain).GetScan().GetTxInfoByHash(txHash)
	if err != nil {
		return nil, err
	}
	if err != nil {
		log.Error(err.Error())
	}
	x, _ := new(big.Int).SetString(txinfo.Result.Value, 0)
	return &model.TxInfo{
		Token:       "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
		FromAddress: txinfo.Result.From,
		ToAddress:   txinfo.Result.To,
		Data:        x,
	}, nil
}

func (p *Provider) GetTxFromAddress(txHash string, blockNumber uint64, index uint64) (ret string, err error) {
	tx, err := p.scan.GetTxInfoByHash(txHash)
	ret = tx.Result.From
	if err != nil || len(ret) == 0 {
		ret, err = p.chainbase.GetTxSender(txHash)
		if err != nil || len(ret) == 0 {
			ret, err = p.geth.GetTxSender(txHash, blockNumber, uint(index))
		}
	}
	return strings.ToLower(ret), err
}

func (p *Provider) GetContractDeployer(address string) (string, error) {
	deployer, err := p.chainbase.GetContractDeployer(address)
	return deployer, err
}

func (p *Provider) GetLogWithHash(topics0 []string, hash string) (model.Events, error) {
	return p.chainbase.GetLogWithHash(topics0, hash)
}
