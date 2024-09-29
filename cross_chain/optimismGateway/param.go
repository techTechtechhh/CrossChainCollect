package optimismGateway

const (
	//RelayedMessage (index_topic_1 bytes32 msgHash)
	RelayedMessage = "0x4641df4a962071e12719d8c8c8e5ac7fc4d97b927346a3d7a335b1f7517e133c"
	//DepositFinalized (index_topic_1 address l1Token, index_topic_2 address l2Token, index_topic_3 address from, address to, uint256 amount, bytes extraData)
	DepositFinalized = "0xb0444523268717a02698be47d0803aa7468c00acbed2f8bd93a0459cde61dd89"
	//MessagePassed (index_topic_1 uint256 nonce, index_topic_2 address sender, index_topic_3 address target, uint256 value, uint256 gasLimit, bytes data, bytes32 withdrawalHash)
	MessagePassed = "0x02a52367d10742d8032712c1bb8e0144ff1ec5ffda1ed7d70bb05a2744955054"
	//SentMessage (index_topic_1 address target, address sender, bytes message, uint256 messageNonce, uint256 gasLimit)
	SentMessage = "0xcb0f7ffd78f9aee47a248fae8db181db6eee833039123e026dcbff529522e52a"
	//WithdrawalInitiated (index_topic_1 address _l1Token, index_topic_2 address _l2Token, index_topic_3 address _from, address _to, uint256 _amount, bytes _data)
	WithdrawalInitiated = "0x73d170910aba9e6d50b102db522b1dbcd796216f5128b445aa2135272886497e"

	//SentMessageExtension1 (index_topic_1 address sender, uint256 value)
	SentMessageExtension1 = "0x8ebb2ec2465bdb2a06a66fc37a0963af8a2a6a1479d81d56fdb8cbb98096d546"
	//ETHDepositInitiated (index_topic_1 address from, index_topic_2 address to, uint256 amount, bytes extraData)
	ETHDepositInitiated = "0x35d79ab81f2b2017e19afb5c5571778877782d7a8786f5907f93b0f4702f4f23"
	//ERC20DepositInitiated (index_topic_1 address _l1Token, index_topic_2 address _l2Token, index_topic_3 address _from, address _to, uint256 _amount, bytes _data)
	ERC20DepositInitiated = "0x718594027abd4eaed59f95162563e0cc6d0e8d5b86b1c7be8b1b0ac3343d0396"
	//ERC20WithdrawalFinalized (index_topic_1 address _l1Token, index_topic_2 address _l2Token, index_topic_3 address _from, address _to, uint256 _amount, bytes _data)
	ERC20WithdrawalFinalized = "0x3ceee06c1e37648fcbb6ed52e17b3e1f275a1f8c7b22a84b2b84732431e046b3"
	//ETHWithdrawalFinalized (index_topic_1 address from, index_topic_2 address to, uint256 amount, bytes extraData)
	ETHWithdrawalFinalized = "0x2ac69ee804d9a7a0984249f508dfab7cb2534b465b6ce1580f99a38ba9c5e631"
)

var OptiContracts = map[string][]string{
	"eth": {
		"0x25ace71c97b33cc4729cf772ae268934f7ab5fa1",
		"0x99c9fc46f92e8a1c0dec1b1747d010903e884be1",
	},
	"optimism": {
		"0x4200000000000000000000000000000000000010",
		"0x4200000000000000000000000000000000000007",
	},
}
