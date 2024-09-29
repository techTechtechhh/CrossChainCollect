package utils

import "math/big"

// define unstandard chanid (btc solana ...)

const (
	BTCChainId = iota + 100000000
	SolanaChainId
	NearChainId
	TerraColumbus
	TerraPhoenix
	Aptos
	Switcheo
	Heimdall
)

var unstandardChains = map[string]*big.Int{
	"eth":           new(big.Int).SetUint64(1),
	"bsc":           new(big.Int).SetUint64(56),
	"polygon":       new(big.Int).SetUint64(137),
	"avalanche":     new(big.Int).SetUint64(43114),
	"arbitrum":      new(big.Int).SetUint64(42161),
	"arbitrum-nova": new(big.Int).SetUint64(42170),
	"xDai":          new(big.Int).SetUint64(200),
	"optimism":      new(big.Int).SetUint64(10),
	"cronos":        new(big.Int).SetUint64(25),
	"fantom":        new(big.Int).SetUint64(250),
	"moonbeam":      new(big.Int).SetUint64(1284),
	"ontology":      new(big.Int).SetUint64(58),
	"neo":           new(big.Int).SetUint64(259),
	"heco":          new(big.Int).SetUint64(128),
	"palette":       new(big.Int).SetUint64(1718),
	"zilliqa":       new(big.Int).SetUint64(32769),
	"curve":         new(big.Int).SetUint64(827431),
	"okx":           new(big.Int).SetUint64(66),
	"gnosis":        new(big.Int).SetUint64(100),
	"metis":         new(big.Int).SetUint64(1088),
	"boba":          new(big.Int).SetUint64(288),
	"oasis":         new(big.Int).SetUint64(42262),
	"harmony":       new(big.Int).SetUint64(1666600000),
	"hsc":           new(big.Int).SetUint64(70),
	"kcc":           new(big.Int).SetUint64(321),
	"kava":          new(big.Int).SetUint64(2222),
	"cube":          new(big.Int).SetUint64(1818),
	"celo":          new(big.Int).SetUint64(42220),
	"astar":         new(big.Int).SetUint64(592),

	// non-standard chain id
	"btc":              new(big.Int).SetUint64(BTCChainId),
	"solana":           new(big.Int).SetUint64(SolanaChainId),
	"near":             new(big.Int).SetUint64(NearChainId),
	"terra-columbus-5": new(big.Int).SetUint64(TerraColumbus),
	"terra-phoenix-1":  new(big.Int).SetUint64(TerraPhoenix),
	"aptos":            new(big.Int).SetUint64(Aptos),
	"switcheo":         new(big.Int).SetUint64(Switcheo),
	"heimdall":         new(big.Int).SetUint64(Heimdall),
}

func GetChainId(name string) *big.Int {
	if val, ok := unstandardChains[name]; ok {
		return new(big.Int).Set(val)
	}
	return new(big.Int).SetUint64(0)
}
