package model

type Collector interface {
	Name() string
}

type EventCollector interface {
	Collector
	SrcTopics0() []string
	Topics0(chain string) []string
	Extract(chain string, events Events) Results
	Contracts(chain string) map[string]string
}

type MsgCollector interface {
	Collector
	Selectors(chain string) []string
	Extract(chain string, msgs []*Call) Results
	Contracts(chain string) []string
}

type TransferCollector interface {
	Collector
	Extract(chain string, msg ERC20Transfers) Results
	Addresses(chain string) []string
}
