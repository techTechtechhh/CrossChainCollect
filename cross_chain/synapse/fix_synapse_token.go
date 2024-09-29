package synapse

import (
	"app/dao"
	"app/model"
	"app/provider"
	"app/svc"
	"app/utils"
	"bytes"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
	"math/big"
	"sort"
	"strings"
	"time"
)

var nUSD = map[string]string{
	"eth":       "0x1b84765de8b7566e4ceaf4d0fd3c5af52d3dde4f",
	"bsc":       "0x23b891e5c62e0955ae2bd185990103928ab817b3",
	"fantom":    "0xED2a7edd7413021d440b09D654f3b87712abAB66",
	"optimism":  "0x809DC529f07651bD43A172e8dB6f4a7a0d771036",
	"polygon":   "0xB6c473756050dE474286bED418B77Aeac39B02aF",
	"arbitrum":  "0x3ea9B0ab55F34Fb188824Ee288CeaEfC63cf908e",
	"avalanche": "0xCFc37A6AB183dd4aED08C204D1c2773c0b1BDf46",
}

var swapPool = map[string]string{
	"bsc":       "0x28ec0b36f0819ecb5005cab836f4ed5a2eca4d13",
	"optimism":  "0xe27bff97ce92c3e1ff7aa9f86781fdd6d48f5ee9",
	"avalanche": "0xed2a7edd7413021d440b09d654f3b87712abab66",
	"fantom":    "0x85662fd123280827e11c59973ac9fcbe838dc3b4",
	"eth":       "0x1116898dda4015ed8ddefb84b6e8bc24528af2d8",
	"polygon":   "0x85fcd7dd0a1e1a9fcd5fd886ed522de8221c3ee5",
	"arbitrum":  "0xa067668661c84476afcdc6fa5d758c4c01c34352",
}

var indexToken = map[string]map[int64]string{
	"0xa067668661c84476afcdc6fa5d758c4c01c34352": {
		0: "0x3ea9b0ab55f34fb188824ee288ceaefc63cf908e",
		1: "0x82af49447d8a07e3bd95bd0d56f35241523fbab1",
	},
	"0x28ec0b36f0819ecb5005cab836f4ed5a2eca4d13": {
		0: "0x23b891e5c62e0955ae2bd185990103928ab817b3",
		1: "0xe9e7cea3dedca5984780bafc599bd69add087d56",
		2: "0x8ac76a51cc950d9822d68b83fe1ad97b32cd580d",
		3: "0x55d398326f99059ff775485246999027b3197955",
	},
	"0xe27bff97ce92c3e1ff7aa9f86781fdd6d48f5ee9": {
		0: "0x809dc529f07651bd43a172e8db6f4a7a0d771036",
		1: "0x121ab82b49b2bc4c7901ca46b8277962b4350204",
	},
	"0xed2a7edd7413021d440b09d654f3b87712abab66": {
		0: "0xcfc37a6ab183dd4aed08c204d1c2773c0b1bdf46",
		1: "0xd586e7f844cea2f87f50152665bcbc2c279d8d70",
		2: "0xa7d7079b0fead91f3e65f86e8915cb59c1a4c664",
		3: "0xc7198437980c041c805a1edcba50c1ce5db95118",
	},
	"0x85662fd123280827e11c59973ac9fcbe838dc3b4": {
		0: "0xed2a7edd7413021d440b09d654f3b87712abab66",
		1: "0x04068da6c83afcfa0e13ba15a6696662335d5b75",
		2: "0x049d68029688eabf473097a2fc38ef61633a3c7a",
	},
	"0x1116898dda4015ed8ddefb84b6e8bc24528af2d8": {
		0: "0x6b175474e89094c44da98b954eedeac495271d0f",
		1: "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48",
		2: "0xdac17f958d2ee523a2206206994597c13d831ec7",
	},
	"0x85fcd7dd0a1e1a9fcd5fd886ed522de8221c3ee5": {
		0: "0xb6c473756050de474286bed418b77aeac39b02af",
		1: "0x8f3cf7ad23cd3cadbd9735aff958023239c6a063",
		2: "0x2791bca1f2de4661ed88a30c99a7a9449aa84174",
		3: "0xc2132d05d31c914a87c6611c10748aeb04b58e8f",
	},
}

func initSynapse() {
	for k, v := range nUSD {
		nUSD[k] = strings.ToLower(v)
	}
	for k, v := range swapPool {
		swapPool[k] = strings.ToLower(v)
	}
	for k, index := range indexToken {
		for idx, token := range index {
			indexToken[k][idx] = strings.ToLower(token)
		}
	}
	/*for chain, poolAddr := range swapPool {
		getPoolToken(chain, poolAddr)
	}*/
}

const (
	TokenSwap       = "0xc6c1e0630dbe9130cc068028486c0d118ddcea348550819defd5cb8c257f8a38"
	RemoveLiquidity = "0x43fb02998f4e03da2e0e6fff53fdbf0c40a9f45f145dc377fc30615d7d7a8a64"
	AddLiquidity    = "0x189c623b666b1b45b83d7178f39b8c087cb09774317ca2f53c2d3c3726f222a2"
	outRangeErr     = "execution reverted: Out of range"
)

func getSynapseLogs(provider *provider.Provider, topics0 []string, hash string) model.Events {
	swapEvent, err := provider.GetLogWithHash(topics0, hash)
	var i = 0
	for err != nil && i < 6 {
		i++
		time.Sleep(10 * time.Second)
		swapEvent, err = provider.GetLogWithHash([]string{TokenSwap, RemoveLiquidity, AddLiquidity}, hash)
	}
	if i >= 6 || err != nil {
		fmt.Println(hash, err)
		return nil
	}
	return swapEvent
}

func updateToken(stmt string, dao *dao.Dao) {
	_, err := dao.DB().Exec(stmt)
	if err != nil {
		fmt.Println(err)
	}
}

func extractRealToken(r *model.Result, events model.Events) string {
	for i := len(events) - 1; i >= 0; i-- {
		e := events[i]
		if e.Id >= r.ActionId {
			continue
		}
		pool := swapPool[r.Chain]
		switch e.Topics[0] {
		case TokenSwap:
			soldId, _ := new(big.Int).SetString(e.Data[len(e.Data)-64*2:len(e.Data)-64], 16)
			boughtId, _ := new(big.Int).SetString(e.Data[len(e.Data)-64:], 16)
			if r.Token == indexToken[pool][soldId.Int64()] {
				return indexToken[pool][boughtId.Int64()]
			} else if r.Token == indexToken[pool][boughtId.Int64()] {
				return indexToken[pool][soldId.Int64()]
			} else {
				continue
			}
		case AddLiquidity:
			a, err := abi.JSON(bytes.NewBufferString(addLiquidityEvent))
			ev, err := a.EventByID(common.HexToHash(AddLiquidity))
			if err != nil {
				log.Error("Synapse Exact() can't decode AddLiquidity", "Chain", r.Chain, "Hash", e.Hash)
			}
			ss, err := ev.Inputs.Unpack(hexutil.MustDecode(e.Data))
			array := ss[0].([]*big.Int)
			for i, amount := range array {
				if amount.Int64() != 0 {
					return indexToken[pool][int64(i)]
				}
			}
		case RemoveLiquidity:
			boughtId, _ := new(big.Int).SetString(e.Data[len(e.Data)-64*2:len(e.Data)-64], 16)
			return indexToken[pool][boughtId.Int64()]
		}
	}
	return ""
}

func getPoolToken(chain, pool string) {
	indexToken[pool] = make(map[int64]string)
	for i := 0; ; i++ {
		res, err := utils.QueryGethWithAbi(chain, pool, getTokenAbi, "getToken", uint8(i))
		if err != nil && err.Error() != outRangeErr {
			fmt.Println(chain, pool, i, err)
		} else if err != nil && err.Error() == outRangeErr {
			return
		} else {
			indexToken[pool][int64(i)] = strings.ToLower(res.(common.Address).String())
		}
	}
}

func getSwapEvent(chain, token string, scx *svc.ServiceContext) {
	defer scx.Wg.Done()
	provider := scx.Providers.Get(chain)
	s := fmt.Sprintf("select * from %s_%s where project = 'Synapse' and token = '%s' and ts >= '2023-05-01'", scx.Dao.Table(), chain, token)
	var results model.Results
	err := scx.Dao.DB().Select(&results, s)
	if err != nil {
		fmt.Println(err)
	}
	var swapAddr = make(map[string]struct{})
	for _, r := range results {
		swapEvent, err := provider.GetLogWithHash([]string{TokenSwap, RemoveLiquidity, AddLiquidity}, r.Hash)
		var i = 0
		for err != nil && i < 6 {
			i++
			time.Sleep(10 * time.Second)
			swapEvent, err = provider.GetLogWithHash([]string{TokenSwap, RemoveLiquidity, AddLiquidity}, r.Hash)
		}
		if i >= 6 || err != nil {
			fmt.Println(r.Hash, r.Chain, r.Direction)
			fmt.Println(err)
			continue
		} else if swapEvent != nil {
			swapAddr[swapEvent[0].Address] = struct{}{}
		}
	}
	fmt.Println(chain)
	for addr := range swapAddr {
		ss := fmt.Sprintf("\"%s\"", addr)
		fmt.Println(ss)
	}
}

var getTokenAbi = `[
{
	"inputs": [{
		"internalType": "uint8",
		"name": "index",
		"type": "uint8"
	}],
	"name": "getToken",
	"outputs": [{
		"internalType": "contract IERC20",
		"name": "",
		"type": "address"
	}],
	"stateMutability": "view",
	"type": "function"
}]`

var addLiquidityEvent = `
[{
	"anonymous": false,
	"inputs": [{
		"indexed": true,
		"internalType": "address",
		"name": "provider",
		"type": "address"
	}, {
		"indexed": false,
		"internalType": "uint256[]",
		"name": "tokenAmounts",
		"type": "uint256[]"
	}, {
		"indexed": false,
		"internalType": "uint256[]",
		"name": "fees",
		"type": "uint256[]"
	}, {
		"indexed": false,
		"internalType": "uint256",
		"name": "invariant",
		"type": "uint256"
	}, {
		"indexed": false,
		"internalType": "uint256",
		"name": "lpTokenSupply",
		"type": "uint256"
	}],
	"name": "AddLiquidity",
	"type": "event"
}]`

//修synapse历史数据

func fixSynapseToken(scx *svc.ServiceContext) {
	initSynapse()
	for chain := range swapPool {
		scx.Wg.Add(1)
		go func(chain string) {
			replaySynapseTx(chain, scx)
		}(chain)
	}
	scx.Wg.Wait()
}
func replaySynapseTx(chain string, scx *svc.ServiceContext) {
	defer scx.Wg.Done()
	provider := scx.Providers.Get(chain)
	token := nUSD[chain]
	s := fmt.Sprintf("select * from %s_%s where project = 'Synapse' and token = '%s' and ts >= '2023-08-01'", scx.Dao.Table(), chain, token)
	var results model.Results
	err := scx.Dao.DB().Select(&results, s)
	if err != nil {
		fmt.Println(err)
	}
	var stmt = "update %s_%s set token = '%s' where hash = '%s'"
	for _, r := range results {
		swapEvents := getSynapseLogs(provider, []string{TokenSwap, RemoveLiquidity, AddLiquidity}, r.Hash)
		if swapEvents == nil {
			continue
		}
		sort.Sort(swapEvents)
		realToken := extractRealToken(r, swapEvents)
		if realToken == "" {
			fmt.Println(chain, r.Hash)
			continue
		}
		var st = fmt.Sprintf(stmt, scx.Dao.Table(), chain, realToken, r.Hash)
		updateToken(st, scx.Dao)
	}
}
