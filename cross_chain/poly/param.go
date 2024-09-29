package poly

import "app/utils"

const (
	CrossChainEvent               = "0x6ad3bf15c1988bc04bc153490cab16db8efb9a3990215bf1c64ea6e28be88483"
	VerifyHeaderAndExecuteTxEvent = "0x8a4a2663ce60ce4955c595da2894de0415240f1ace024cfbff85f513b656bdae"
	LockEvent                     = "0x8636abd6d0e464fe725a13346c7ac779b73561c705506044a2e6b2cdb1295ea5"
	UnLockEvent                   = "0xd90288730b87c2b8e0c45bd82260fd22478aba30ae1c4d578b8daba9261604df"
	UnLockEvent_Switcheo          = "0x2d3f6ad356f1c408166244c68a928a722472299760d71a6de97f6057b912972c"
	LockEvent_Switcheo            = "0x3aa1a37a3bb16943a2c97dd810c5601a4ce19bb1942a54401f821af5515c5530"
	SendMsg                       = "0x8d3ee0df6a4b7e82a7f20a763f1c6826e6176323e655af64f32318827d2112d4"
)

var PolyContracts = map[string][]string{
	"bsc": {
		"0x1c9ca8abb5da65d94dad2e8fb3f45535480d5909", //crosschain event
		"0x2f7ac9436ba4b548f9582af91ca1ef02cd2f1f03", //查lockEvent
		"0x1c9ca8abb5da65d94dad2e8fb3f45535480d5909", //switcheo
	},
	"eth": {
		"0x14413419452aaf089762a0c5e95ed2a13bbc488c", //ethCrossChainManager
		"0x250e76987d838a75310c34bf422ea9f1ac4cc906", //Bridge
		"0x9a016Ce184a22DbF6c17daA59Eb7d3140DBd1c54", //switcheo
	},
	"polygon": {
		"0x5f8517d606580d30c3bf210fa016b8916c685be8", //Bridge
		"0xb16fed79a6cb9270956f045f2e7989affb75d459", //ethCrossChainManager
		"0x43138036d1283413035b8eca403559737e8f7980", //switcheo
	},
	"arbitrum": {
		"0x30e39786f0dd700da277a54bd9c07f7894cb5aba",
		"0x7cea671dabfba880af6723bddd6b9f4caa15c87b", //manager
		"0xb1e6f8820826491fcc5519f84ff4e2bdbb6e3cad", //switcheo
	},
	"avalanche": {
		"0xd3b90e2603d265bf46dbc788059ac12d52b6ac57", //bridge
		"0x2aa63cd0b28fb4c31fa8e4e95ec11815be07b9ac", //manager
	},
	"fantom": {
		"0xd3b90e2603d265bf46dbc788059ac12d52b6ac57", //bridge
		//"0x7bb9709ec786ea549ee67ae02e8b0c75dde77f48", //bridge
		"0x2aa63cd0b28fb4c31fa8e4e95ec11815be07b9ac", //manager
	},
	"optimism": {
		"0x2aa63cd0b28fb4c31fa8e4e95ec11815be07b9ac", //manager
		"0x8a05dc902d15aea923f2c722292f5561c3496317", //bridge
	},
}

func init() {
	for name, chain := range PolyContracts {
		PolyContracts[name] = utils.StrSliceToLower(chain)
	}
}

type Detail struct {
	CrossChainTxHash string `json:"cross_chain_tx_hash,omitempty"`
	ToContract       string `json:"to_contract,omitempty"`
	ToAsset          string `json:"to_asset,omitempty"`
}
