package avaxBridge

const (
	//Mint (address to, uint256 amount, address feeAddress, uint256 feeAmount, bytes32 originTxId)
	Mint = "0x918d77674bb88eaf75afb307c9723ea6037706de68d6fc07dd0c6cba423a5250"
	//Mint = "0x67fc19bb"
	//Transfer (index_topic_1 address from, index_topic_2 address to, uint256 value)
	Transfer = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
	//Transfer = "0xa9059cbb"
)

const avaxFeeCollector = "0x6283184a580ec470fed64f75a20edfe4917f9ffe"

var AvaxBridgeContracts = map[string]string{
	"eth":       "0x8eb8a3b98659cce290402893d0123abb75e3ab28", //eth上的avalanche bridge
	"avalanche": "0xeb1bb70123b2f43419d070d7fde5618971cc2f8f",
}

var matchChain = map[string]string{
	"eth":       "avalanche",
	"avalanche": "eth",
}
