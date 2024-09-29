package arbitrumBridge

import "app/utils"

const (
	DepositInitiated          = "0xb8910b9960c443aac3240b98585384e3a6f109fbf6969e264c3f183d69aba7e1" //eth上跨出
	TxToL2                    = "0xc1d1490cf25c3b40d600dfb27c7680340ed1ab901b7e8f3551280968a3b372b0"
	OutBoxTransactionExecuted = "0x20af7f3bbfe38132b8900ae295cd9c8d1914be7052d061a511f3f728dab18964"
	WithdrawalFinalized       = "0x891afe029c75c4f8c5855fc3480598bc5a53739344f6ae575bdb7ea2a79f56b3" //eth上跨入

	TxToL1              = "0x2b986d32a0536b7e19baa48ab949fec7b903b7fad7730820b20632d100cc3a68"
	WithdrawalInitiated = "0x3073a74ecb728d10be779fe19a74a1428e20468f5b4d167bf9c73d9067847d73" //arbi上跨出
	DepositFinalized    = "0xc7f2e9c55c40a50fbc217dfc70cd39a222940dfa62145aa0ca49eb9535d4fcb2" //arbi上跨入
)

var OtherBridgeContractOnETH = map[string]string{
	"0xb2535b988dce19f9d71dfb22db6da744acac21bf": "arbitrum-nova",
	"0x23122da8c581aa7e0d07a36ff1f16f799650232f": "arbitrum-nova",
	"0x97f63339374fce157aa8ee27830172d2af76a786": "xDai", //名字必须跟utils.GetChainId函数里的一致
}

var ArbiContracts = map[string][]string{
	"eth": {
		"0x97f63339374fce157aa8ee27830172d2af76a786", //xDai
		"0xe4e2121b479017955be0b175305b35f312330bae",
		"0x97f63339374fce157aa8ee27830172d2af76a786",
		"0xe4e2121b479017955be0b175305b35f312330bae",
		"0x23122da8c581aa7e0d07a36ff1f16f799650232f", //arbitrum nova bridge
		"0x797758746c150cdcaac38d294319322d2b753d4a",
		"0xb2535b988dce19f9d71dfb22db6da744acac21bf", //arbitrum nova bridge
		"0xbbce8aa77782f13d4202a230d978f361b011db27",
		"0x6142f1c8bbf02e6a6bd074e8d564c9a5420a0676",
		"0x01cdc91b0a9ba741903aa3699bf4ce31d6c5cc06",
		"0xd3b5b60020504bc3489d6949d545893982ba3011",
		"0xd92023e9d9911199a6711321d1277285e6d4e2db",
		"0x0f25c1dc2a9922304f2eac71dca9b07e310e8e5a",
		"0xcee284f754e854890e311e3280b767f80797180d",
		"0xa3a7b6f88361f48403514059f1f16c8e78d60eec",
	},
	"arbitrum": {
		"0x65e1a5e8946e7e87d9774f5288f41c30a99fd302",
		"0x05d2218b4586a78785fdca7b92322d3293c82eee",
		"0x6d2457a4ad276000a615295f7a80f79e48ccd318",
		"0x07d4692291b9e30e326fd31706f686f83f331b82",
		"0x6c411ad3e74de3e7bd422b94a27770f5b86c623b",
		"0x467194771dae2967aef3ecbedd3bf9a310c76c65",
		"0xcad7828a19b363a2b44717afb1786b5196974d8e",
		"0x09e9222e96e7b4ae2a407b98d48e330053351eee",
		"0x096760f208390250649e3e8763348e783aef5562",
	},
}

func init() {
	for name, chain := range ArbiContracts {
		ArbiContracts[name] = utils.StrSliceToLower(chain)
	}
}

type Detail struct {
	l1Token string `json:"l1_token,omitempty"`
}
