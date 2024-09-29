package model

import (
	"math/big"
	"time"
)

type EventConfig struct {
	Addresses []string
	Topics0   []string
}

type MsgConfig struct {
	Addresses []string
	Selector  []string
}

type Event struct {
	Number  uint64    `json:"number"`
	Ts      time.Time `json:"ts"`
	Index   uint64    `json:"index"`
	Hash    string    `json:"hash"`
	Id      uint64    `json:"id"`
	Address string    `json:"address"`
	Topics  []string  `json:"topics"`
	Data    string    `json:"data"`
}

// 用于filler
type TransferInfo struct {
	Token       string
	FromAddress string
	ToAddress   string
	Data        string
}

type TxInfo struct {
	Token       string
	FromAddress string
	ToAddress   string
	Data        *big.Int
}

type Events []*Event

func (e Events) Len() int {
	return len(e)
}

func (e Events) Less(i, j int) bool {
	if e[i].Number != e[j].Number {
		return e[i].Number < e[j].Number
	} else if e[i].Index != e[j].Index {
		return e[i].Index < e[j].Index
	}
	return e[i].Id < e[j].Id
}

func (e Events) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

type Call struct {
	Number uint64    `json:"number"`
	Ts     time.Time `json:"ts"`
	Index  uint64    `json:"index"`
	Hash   string    `json:"hash"`
	Id     uint64    `json:"id"`
	From   string    `json:"from"`
	To     string    `json:"to"`
	Input  string    `json:"input"`
	Value  *big.Int  `json:"value"`
}

type ERC20TransferInfo struct {
	Number          uint64
	Ts              time.Time
	From            string
	To              string
	Hash            string
	ContractAddress string
	Value           *big.Int
	Index           uint64
	ActionId        uint64
	Input           string
}

type ERC20Transfers []*ERC20TransferInfo

func (e ERC20Transfers) Len() int {
	return len(e)
}

func (e ERC20Transfers) Less(i, j int) bool {
	if e[i].Number != e[j].Number {
		return e[i].Number < e[j].Number
	} else if e[i].Index != e[j].Index {
		return e[i].Index < e[j].Index
	}
	return e[i].ActionId < e[j].ActionId
}

func (e ERC20Transfers) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}
