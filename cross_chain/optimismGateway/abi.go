package optimismGateway

var L2CrossDomainMessenger = map[string]string{
	"optimism": `[{
	"anonymous": false,
	"inputs": [{
		"indexed": true,
		"internalType": "address",
		"name": "target",
		"type": "address"
	}, {
		"indexed": false,
		"internalType": "address",
		"name": "sender",
		"type": "address"
	}, {
		"indexed": false,
		"internalType": "bytes",
		"name": "message",
		"type": "bytes"
	}, {
		"indexed": false,
		"internalType": "uint256",
		"name": "messageNonce",
		"type": "uint256"
	}, {
		"indexed": false,
		"internalType": "uint256",
		"name": "gasLimit",
		"type": "uint256"
	}],
	"name": "SentMessage",
	"type": "event"
},{
	"inputs": [{
		"internalType": "uint256",
		"name": "_nonce",
		"type": "uint256"
	}, {
		"internalType": "address",
		"name": "_sender",
		"type": "address"
	}, {
		"internalType": "address",
		"name": "_target",
		"type": "address"
	}, {
		"internalType": "uint256",
		"name": "_value",
		"type": "uint256"
	}, {
		"internalType": "uint256",
		"name": "_minGasLimit",
		"type": "uint256"
	}, {
		"internalType": "bytes",
		"name": "_message",
		"type": "bytes"
	}],
	"name": "relayMessage",
	"outputs": [],
	"stateMutability": "payable",
	"type": "function"
}]`,
}

var L2StandardBridge = map[string]string{
	"optimism": `[{
	"anonymous": false,
	"inputs": [{
		"indexed": true,
		"internalType": "address",
		"name": "l1Token",
		"type": "address"
	}, {
		"indexed": true,
		"internalType": "address",
		"name": "l2Token",
		"type": "address"
	}, {
		"indexed": true,
		"internalType": "address",
		"name": "from",
		"type": "address"
	}, {
		"indexed": false,
		"internalType": "address",
		"name": "to",
		"type": "address"
	}, {
		"indexed": false,
		"internalType": "uint256",
		"name": "amount",
		"type": "uint256"
	}, {
		"indexed": false,
		"internalType": "bytes",
		"name": "extraData",
		"type": "bytes"
	}],
	"name": "DepositFinalized",
	"type": "event"
}, {
	"anonymous": false,
	"inputs": [{
		"indexed": true,
		"internalType": "address",
		"name": "l1Token",
		"type": "address"
	}, {
		"indexed": true,
		"internalType": "address",
		"name": "l2Token",
		"type": "address"
	}, {
		"indexed": true,
		"internalType": "address",
		"name": "from",
		"type": "address"
	}, {
		"indexed": false,
		"internalType": "address",
		"name": "to",
		"type": "address"
	}, {
		"indexed": false,
		"internalType": "uint256",
		"name": "amount",
		"type": "uint256"
	}, {
		"indexed": false,
		"internalType": "bytes",
		"name": "extraData",
		"type": "bytes"
	}],
	"name": "WithdrawalInitiated",
	"type": "event"
}]`,
}

var L2ToL1MessagePasser = map[string]string{
	"optimism": `[{
	"inputs": [],
	"stateMutability": "nonpayable",
	"type": "constructor"
}, {
	"anonymous": false,
	"inputs": [{
		"indexed": true,
		"internalType": "uint256",
		"name": "nonce",
		"type": "uint256"
	}, {
		"indexed": true,
		"internalType": "address",
		"name": "sender",
		"type": "address"
	}, {
		"indexed": true,
		"internalType": "address",
		"name": "target",
		"type": "address"
	}, {
		"indexed": false,
		"internalType": "uint256",
		"name": "value",
		"type": "uint256"
	}, {
		"indexed": false,
		"internalType": "uint256",
		"name": "gasLimit",
		"type": "uint256"
	}, {
		"indexed": false,
		"internalType": "bytes",
		"name": "data",
		"type": "bytes"
	}, {
		"indexed": false,
		"internalType": "bytes32",
		"name": "withdrawalHash",
		"type": "bytes32"
	}],
	"name": "MessagePassed",
	"type": "event"
}]`,
}

var L1StandardBridge = map[string]string{
	"eth": `[{
	"anonymous": false,
	"inputs": [{
		"indexed": true,
		"internalType": "address",
		"name": "from",
		"type": "address"
	}, {
		"indexed": true,
		"internalType": "address",
		"name": "to",
		"type": "address"
	}, {
		"indexed": false,
		"internalType": "uint256",
		"name": "amount",
		"type": "uint256"
	}, {
		"indexed": false,
		"internalType": "bytes",
		"name": "extraData",
		"type": "bytes"
	}],
	"name": "ETHDepositInitiated",
	"type": "event"
}]`,
}

var L1CrossDomainMessenger = map[string]string{
	"eth": `[{
	"inputs": [{
		"internalType": "address",
		"name": "_target",
		"type": "address"
	}, {
		"internalType": "address",
		"name": "_sender",
		"type": "address"
	}, {
		"internalType": "bytes",
		"name": "_message",
		"type": "bytes"
	}, {
		"internalType": "uint256",
		"name": "_nonce",
		"type": "uint256"
	}],
	"name": "relayMessage",
	"outputs": [],
	"stateMutability": "payable",
	"type": "function"
}]`,
}
