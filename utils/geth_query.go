package utils

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"golang.org/x/exp/slices"
	"strings"
)

func QueryGethWithAbi(chain, contract, rawAbi, funcName string, args ...interface{}) (interface{}, error) {
	var err error
	// 连接RPC节点
	urls := rpcNodes[chain]
	url := urls[0]
	client, err := ethclient.Dial(url)
	if err != nil {
		fmt.Println("无法连接到RPC节点：", url, err)
		return nil, err
	}

	// 要查询的合约地址
	contractAddress := common.HexToAddress(contract)

	contractAbi, err := abi.JSON(strings.NewReader(rawAbi))
	if err != nil {
		return nil, err
	}

	// 构造函数调用数据
	data, err := contractAbi.Pack(funcName, args...)
	if err != nil {
		return nil, err
	}

	// 调用合约的视图函数
	callData := ethereum.CallMsg{
		To:   &contractAddress,
		Data: data,
	}
	result, err := client.CallContract(context.Background(), callData, nil)
	if err != nil {
		return nil, err
	}

	// 解析合约返回结果
	var token interface{}
	err = contractAbi.UnpackIntoInterface(&token, funcName, result)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func QueryGeth(chain, address, sig string) interface{} {
	var err error
	// 连接RPC节点
	urls := rpcNodes[chain]
	url := urls[0]
	client, err := ethclient.Dial(url)
	if err != nil {
		fmt.Println("无法连接到RPC节点：", url, err)
		return -1
	}

	// 要查询的合约地址
	contractAddress := common.HexToAddress(address)

	// 构造调用消息
	input, err := hexToBytes(sig)
	if err != nil {
		fmt.Println("消息构造失败：", err)
		return -1
	}

	// 构造调用请求
	callMsg := ethereum.CallMsg{
		To:   &contractAddress,
		Data: input,
	}

	// 执行调用请求
	result, err := client.CallContract(context.Background(), callMsg, nil)
	var try = 0
	for err != nil && err.Error() != "execution reverted" {
		try++
		if try == len(urls)*2 {
			log.Error("调用失败：", "Err", err)
			return -1
		}
		client, _ = ethclient.Dial(url)
		url = nextNode(urls, url)
		client, _ = ethclient.Dial(url)
		result, err = client.CallContract(context.Background(), callMsg, nil)
	}
	return result
}

// 将16进制字符串转换为字节数组
func hexToBytes(hex string) ([]byte, error) {
	if len(hex)%2 != 0 {
		return nil, fmt.Errorf("无效的16进制字符串: %s", hex)
	}
	bytes := make([]byte, len(hex)/2)
	for i := 0; i < len(hex)/2; i++ {
		n, err := fmt.Sscanf(hex[2*i:2*i+2], "%02x", &bytes[i])
		if n != 1 || err != nil {
			return nil, fmt.Errorf("无效的16进制字符串: %s", hex)
		}
	}
	return bytes, nil
}

func nextNode(rpcNodes []string, node string) string {
	idx := slices.Index(rpcNodes, node)
	idx++
	idx %= len(rpcNodes)
	return rpcNodes[idx]
}

var rpcNodes = map[string][]string{
	"eth":      {"https://eth-mainnet.g.alchemy.com/v2/OckUvkRX1aiG0hWaCJiK0cV4dcr2xf5t", "https://eth-mainnet.g.alchemy.com/v2/UsIMQMQTPzEJLkiiAUXrq_Zg8O-sl1Ek"},
	"ethereum": {"https://eth-mainnet.g.alchemy.com/v2/OckUvkRX1aiG0hWaCJiK0cV4dcr2xf5t", "https://eth-mainnet.g.alchemy.com/v2/UsIMQMQTPzEJLkiiAUXrq_Zg8O-sl1Ek"},
	"polygon":  {"https://polygon-rpc.com", "https://polygon-mainnet.g.alchemy.com/v2/6VywoiguHDFSI-34a3qZQxPk3P2Q9oxt"},
	"arbitrum": {"https://arb-mainnet.g.alchemy.com/v2/YQ9AU6YZxrJ4pmhybu0eE2JX1m5A4He9", "https://arb1.arbitrum.io/rpc", "https://arb-mainnet.g.alchemy.com/v2/3NlFyO762rEpQru4axc_2az5iNh19D0c"},
	"fantom":   {"https://ftm.getblock.io/0b86cec7-52c8-4a2a-858a-683d28ddccce/mainnet/", "https://rpcapi.fantom.network"},
	"optimism": {"https://opt-mainnet.g.alchemy.com/v2/mpeRTW6iPVW4b7eYYxYWZq0FNQOfSXnX", "https://1rpc.io/op"},
	"avalanche": {"https://ava-mainnet.public.blastapi.io/ext/bc/C/rpc",
		"https://avax.getblock.io/mainnet/0b86cec7-52c8-4a2a-858a-683d28ddccce/ext/bc/C/rpc",
		"https://rpc.ankr.com/avalanche",
		"https://api.avax.network/ext/bc/C/rpc",
		"https://1rpc.io/avax/c",
		"https://1rpc.io/avax/c"},
	"bsc": {"https://bsc-mainnet.nodereal.io/v1/64a9df0874fb4a93b9d0a3849de012d3", "https://rpc.ankr.com/bsc",
		"https://rpc-bsc.bnb48.club", "https://bsc-dataseed1.defibit.io", "https://bsc-dataseed2.defibit.io",
		"https://bsc-dataseed3.defibit.io", "https://bsc-dataseed4.defibit.io", "https://bsc-dataseed1.ninicoin.io",
		"https://bsc-dataseed2.ninicoin.io", "https://bsc-dataseed3.ninicoin.io", "https://bsc-dataseed4.ninicoin.io",
		"https://bscrpc.com", "https://bsc-mainnet.public.blastapi.io", "https://binance.nodereal.io",
		"https://bsc.mytokenpocket.vip"},
}
