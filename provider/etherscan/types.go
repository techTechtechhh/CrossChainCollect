package etherscan

type EtherscanResponse[T any] struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  T      `json:"result"`
}

type GethResponse[T any] struct {
	JsonRpc string `json:"jsonrpc"`
	Id      int    `json:"id"`
	Result  T      `json:"result"`
}

type EtherscanEvent struct {
	BlockNumber string   `json:"blockNumber"`
	Timestamp   string   `json:"timeStamp"`
	Index       string   `json:"transactionIndex"`
	LogIndex    string   `json:"logIndex"`
	Address     string   `json:"address"`
	Hash        string   `json:"transactionHash"`
	Topics      []string `json:"topics"`
	Data        string   `json:"data"`
}

type NormalTx struct {
	BlockNumber     string `json:"blockNumber"`
	Timestamp       string `json:"timeStamp"`
	Hash            string `json:"hash"`
	Index           string `json:"transactionIndex"`
	From            string `json:"from"`
	To              string `json:"to"`
	Value           string `json:"value"`
	Input           string `json:"input"`
	Error           string `json:"isError"`
	ContractAddress string `json:"contractAddress"`
}

type InternalTx struct {
	BlockNumber     string `json:"blockNumber"`
	Timestamp       string `json:"timeStamp"`
	Hash            string `json:"hash"`
	From            string `json:"from"`
	To              string `json:"to"`
	Value           string `json:"value"`
	Input           string `json:"input"`
	ContractAddress string `json:"contractAddress"`
	Error           string `json:"isError"`
	Type            string `json:"type"`
}

type TxInfoByHash struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  struct {
		BlockHash        string `json:"blockHash"`
		BlockNumber      string `json:"blockNumber"`
		From             string `json:"from"`
		Gas              string `json:"gas"`
		GasPrice         string `json:"gasPrice"`
		Hash             string `json:"hash"`
		Input            string `json:"input"`
		Nonce            string `json:"nonce"`
		To               string `json:"to"`
		TransactionIndex string `json:"transactionIndex"`
		Value            string `json:"value"`
		Type             string `json:"type"`
		ChainID          string `json:"chainId"`
		V                string `json:"v"`
		R                string `json:"r"`
		S                string `json:"s"`
	} `json:"result"`
}

type TokenTransfer struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  []struct {
		BlockNumber       string `json:"blockNumber"`
		TimeStamp         string `json:"timeStamp"`
		Hash              string `json:"hash"`
		Nonce             string `json:"nonce"`
		BlockHash         string `json:"blockHash"`
		From              string `json:"from"`
		To                string `json:"to"`
		ContractAddress   string `json:"contractAddress"`
		Value             string `json:"value"`
		TokenName         string `json:"tokenName"`
		TokenSymbol       string `json:"tokenSymbol"`
		TokenDecimal      string `json:"tokenDecimal"`
		TransactionIndex  string `json:"transactionIndex"`
		Gas               string `json:"gas"`
		GasPrice          string `json:"gasPrice"`
		GasUsed           string `json:"gasUsed"`
		CumulativeGasUsed string `json:"cumulativeGasUsed"`
		Input             string `json:"input"`
		Confirmations     string `json:"confirmations"`
	} `json:"result"`
}

type NormalTxOrigin struct {
	BlockNumber       string `json:"blockNumber"`
	TimeStamp         string `json:"timeStamp"`
	Hash              string `json:"hash"`
	Nonce             string `json:"nonce"`
	BlockHash         string `json:"blockHash"`
	TransactionIndex  string `json:"transactionIndex"`
	From              string `json:"from"`
	To                string `json:"to"`
	Value             string `json:"value"`
	Gas               string `json:"gas"`
	GasPrice          string `json:"gasPrice"`
	IsError           string `json:"isError"`
	TxreceiptStatus   string `json:"txreceipt_status"`
	Input             string `json:"input"`
	ContractAddress   string `json:"contractAddress"`
	CumulativeGasUsed string `json:"cumulativeGasUsed"`
	GasUsed           string `json:"gasUsed"`
	Confirmations     string `json:"confirmations"`
	MethodID          string `json:"methodId"`
	FunctionName      string `json:"functionName"`
}
