package model

import (
	"database/sql"
	"github.com/ethereum/go-ethereum/log"
	"reflect"
	"time"
)

type Result struct {
	Id            uint64         `db:"id"`
	MatchId       sql.NullInt64  `db:"match_id"`
	Chain         string         `db:"chain"`
	Number        uint64         `db:"number"`
	Ts            time.Time      `db:"ts"`
	Index         uint64         `db:"index"`
	Hash          string         `db:"hash"`
	ActionId      uint64         `db:"action_id"`
	Project       string         `db:"project"`
	Contract      string         `db:"contract"`
	Direction     string         `db:"direction"`
	TxFromAddress sql.NullString `db:"tx_from_address"`
	FromChainId   *BigInt        `db:"from_chain_id"`
	FromAddress   sql.NullString `db:"from_address"`
	ToChainId     *BigInt        `db:"to_chain_id"`
	ToAddress     sql.NullString `db:"to_address"`
	Token         string         `db:"token"`
	Amount        *BigInt        `db:"amount"`
	MatchTag      string         `db:"match_tag"`
	Detail        []byte         `db:"detail"`
	MatchHash     sql.NullString `db:"match_hash"`
	RealTokenOut  sql.NullString `db:"real_token_out"`
	RealAmountOut *BigInt        `db:"real_amount_out"`
	RealTokenIn   sql.NullString `db:"real_token_in"`
	RealAmountIn  *BigInt        `db:"real_amount_in"`
	Decimals      *int           `db:"decimals"`
}

type Results []*Result

type MatchedId struct {
	SrcID uint64 `db:"src_id"`
	DstID uint64 `db:"dest_id"`
}

type MatchedIds []*MatchedId

func ScanBaseInfo(chain, project string, ele interface{}) (res *Result) {
	switch ele.(type) {
	case *Event:
		e := ele.(*Event)
		res = &Result{
			Chain:    chain,
			Project:  project,
			Number:   e.Number,
			Ts:       e.Ts.UTC(),
			Index:    e.Index,
			Hash:     e.Hash,
			ActionId: e.Id,
			Contract: e.Address,
		}
	case *Call:
		msg := ele.(*Call)
		res = &Result{
			Chain:    chain,
			Number:   msg.Number,
			Ts:       msg.Ts,
			Index:    msg.Index,
			Hash:     msg.Hash,
			ActionId: msg.Id,
			Project:  project,
			Contract: msg.To,
		}
	case *ERC20TransferInfo:
		info := ele.(*ERC20TransferInfo)
		res = &Result{
			Chain:    chain,
			Number:   info.Number,
			Ts:       info.Ts,
			Index:    info.Index,
			Hash:     info.Hash,
			Project:  project,
			Contract: info.ContractAddress,
			ActionId: info.ActionId,
		}
	default:
		log.Error("wrong type", reflect.TypeOf(ele))
		return
	}
	return
}

func (e Results) Len() int {
	return len(e)
}

func (e Results) Less(i, j int) bool {
	if e[i].Number != e[j].Number {
		return e[i].Number < e[j].Number
	} else if e[i].Index != e[j].Index {
		return e[i].Index < e[j].Index
	}
	return e[i].Id < e[j].Id
}

func (e Results) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}
