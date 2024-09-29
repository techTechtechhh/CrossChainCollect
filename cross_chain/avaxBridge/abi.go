package avaxBridge

const avaxAbiStr = `[{
	"constant": false,
	"inputs": [{
		"name": "_to",
		"type": "address"
	}, {
		"name": "_value",
		"type": "uint256"
	}],
	"name": "transfer",
	"outputs": [],
	"payable": false,
	"stateMutability": "nonpayable",
	"type": "function"
}, {
	"inputs": [{
		"internalType": "address",
		"name": "to",
		"type": "address"
	}, {
		"internalType": "uint256",
		"name": "amount",
		"type": "uint256"
	}, {
		"internalType": "address",
		"name": "feeAddress",
		"type": "address"
	}, {
		"internalType": "uint256",
		"name": "feeAmount",
		"type": "uint256"
	}, {
		"internalType": "bytes32",
		"name": "originTxId",
		"type": "bytes32"
	}],
	"name": "mint",
	"outputs": [],
	"stateMutability": "nonpayable",
	"type": "function"
}, {
	"anonymous": false,
	"inputs": [{
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
		"internalType": "address",
		"name": "feeAddress",
		"type": "address"
	}, {
		"indexed": false,
		"internalType": "uint256",
		"name": "feeAmount",
		"type": "uint256"
	}, {
		"indexed": false,
		"internalType": "bytes32",
		"name": "originTxId",
		"type": "bytes32"
	}],
	"name": "Mint",
	"type": "event"
}]`