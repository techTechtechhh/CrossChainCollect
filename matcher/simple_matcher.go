package matcher

import (
	"app/cross_chain/arbitrumBridge"
	"app/dao"
	"app/model"
	"app/svc"
	"app/utils"
	"bytes"
	"database/sql"
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

type SimpleInMatcher struct {
	svc *svc.ServiceContext
	dao *dao.Dao
}

var _ model.Matcher = &SimpleInMatcher{}

func NewSimpleInMatcher(svc *svc.ServiceContext) *SimpleInMatcher {
	return &SimpleInMatcher{dao: svc.Dao, svc: svc}
}

func (m *SimpleInMatcher) Match(crossIns model.Results) (shouldUpdates, unmatches, multis model.Results, err error) {
	if len(crossIns) == 0 {
		return nil, nil, nil, nil
	}
	var matchedCrossOuts = make(map[uint64]struct{})
	sort.Sort(crossIns) //按照时间顺序排序，便于处理multi-match
	//log.Info("simple matcher begins to match", "task", len(crossIns), "from_id", crossIns[0].Id)
	for _, crossIn := range crossIns {
		if crossIn.Direction != model.InDirection {
			log.Warn("matching should not input cross-out")
			continue
		}
		var pending model.Results
		var stmt string
		if crossIn.Project == "Across" {
			stmt = fmt.Sprintf("select * from %s where match_tag = $1 and project = $2 and direction = '%s' and match_id is null and from_chain_id = %s and from_address = '%s' and to_address = '%s' and to_chain_id =  %s", m.dao.Table(), model.OutDirection, crossIn.FromChainId.String(), crossIn.FromAddress.String, crossIn.ToAddress.String, crossIn.ToChainId.String())
		} else if crossIn.FromChainId != nil && crossIn.FromChainId.String() != "" {
			stmt = fmt.Sprintf("select * from %s where match_tag = $1 and project = $2 and direction = '%s' and from_chain_id = %s and match_id is null", m.dao.Table(), model.OutDirection, crossIn.FromChainId.String())
		} else {
			stmt = fmt.Sprintf("select * from %s where match_tag = $1 and project = $2 and direction = '%s' and match_id is null", m.dao.Table(), model.OutDirection)
		}
		err := m.dao.DB().Select(&pending, stmt, crossIn.MatchTag, crossIn.Project)
		if err != nil {
			fmt.Println(err)
			return nil, nil, nil, err
		}
		if len(pending) == 0 {
			unmatches = append(unmatches, crossIn)
			continue
		}
		valid := make(model.Results, 0)
		for _, counterparty := range pending {
			if !IsMatched(counterparty, crossIn) {
				continue
			}
			valid = append(valid, counterparty)
		}
		var matchedOut *model.Result
		var restValids model.Results
		if len(valid) > 1 {
			matchedOut, restValids = DealMulti(crossIn, valid, matchedCrossOuts)
			multis = append(multis, restValids...)
		} else if len(valid) == 1 {
			matchedOut = valid[0]
		} else {
			unmatches = append(unmatches, crossIn)
		}
		if !utils.IsEmpty(matchedOut) {
			shouldUpdates = append(shouldUpdates, matchedOut)
			shouldUpdates = append(shouldUpdates, crossIn)
			matchedCrossOuts[matchedOut.Id] = struct{}{}
			err = m.fillEmptyFields(matchedOut, crossIn)
			if err != nil {
				log.Error("fillEmptyFields failed", "ERR", err)
			}
		}
	}
	return
}

func IsMatched(out, in *model.Result) bool {
	if out.ToChainId != nil && out.ToChainId.Valid() {
		if (*big.Int)(out.ToChainId).Cmp(utils.GetChainId(in.Chain)) != 0 {
			return false
		}
	}
	if in.FromChainId != nil && in.FromChainId.Valid() {
		if (*big.Int)(in.FromChainId).Cmp(utils.GetChainId(out.Chain)) != 0 {
			return false
		}
	}
	if out.FromAddress.Valid && in.FromAddress.Valid && len(out.FromAddress.String) > 0 && len(in.FromAddress.String) > 0 {
		inFromAddr := common.TrimLeftZeroes(common.FromHex(in.FromAddress.String))
		outFromAddr := common.TrimLeftZeroes(common.FromHex(out.FromAddress.String))
		if !bytes.Equal(inFromAddr, outFromAddr) {
			return false
		}
	}
	if out.ToAddress.Valid && in.ToAddress.Valid && len(out.ToAddress.String) > 0 && len(in.ToAddress.String) > 0 {
		inToAddr := common.TrimLeftZeroes(common.FromHex(in.ToAddress.String))
		outToAddr := common.TrimLeftZeroes(common.FromHex(out.ToAddress.String))
		if !bytes.Equal(inToAddr, outToAddr) {
			return false
		}
	}
	return true
}

func DealMulti(crossIn *model.Result, valid model.Results, matchedCrossOuts map[uint64]struct{}) (matchedOut *model.Result, restValids model.Results) {
	if crossIn.Direction != model.InDirection {
		return nil, nil
	}
	if crossIn.Project == arbitrumBridge.NewArbiBridge().Name() && crossIn.Chain == "eth" {
		//因为arbiBridge从L2 -> L1的时间很久，并且需要用户自己申请
		return nil, nil
	}
	sort.Sort(valid)
	restValids = valid
	matchedOut = nil
	for i, v := range valid {
		if v.Direction != model.OutDirection {
			continue
		}
		if _, ok := matchedCrossOuts[v.Id]; ok {
			continue
		}
		//如果已经匹配过了，就跳过
		if v.Ts.After(crossIn.Ts) && i > 0 {
			if _, ok := matchedCrossOuts[valid[i-1].Id]; !ok {
				matchedOut = valid[i-1]
				restValids = append(restValids, valid[:i-1]...)
				restValids = append(restValids, valid[i:]...)
			}
		} else {
			matchedOut = v
			restValids = valid[i+1:]
		}
		break
	}
	return
}

func (m *SimpleInMatcher) fillEmptyFields(out, in *model.Result) error {
	if out == nil || in == nil || out.Direction != model.OutDirection || in.Direction != model.InDirection {
		log.Error("invalid match pair")
		return fmt.Errorf("input format error")
	}

	if in.Id != 0 {
		out.MatchId = sql.NullInt64{Int64: int64(in.Id), Valid: true}
	}
	if out.Id != 0 {
		in.MatchId = sql.NullInt64{Int64: int64(out.Id), Valid: true}
	}

	fromAddress, err := m.svc.Providers.Get(out.Chain).GetTxFromAddress(out.Hash, out.Number, out.Index)
	if out.Project == "Hop" {
		out.FromAddress.Scan(fromAddress)
	}
	out.TxFromAddress.Scan(fromAddress)
	earliest_time, err := time.Parse(model.TIME_LAYOUT, "2019-01-01")
	if err != nil {
		log.Warn("failed parse time", "ERR", err)
	}
	if in.Ts.Before(earliest_time) {
		in.Ts = out.Ts
	}
	if out.Ts.Before(earliest_time) {
		out.Ts = in.Ts
	}

	// fill empty in cross-inc
	if in.FromChainId == nil || !in.FromChainId.Valid() {
		in.FromChainId = (*model.BigInt)(new(big.Int).Set(utils.GetChainId(out.Chain)))
	}
	if !in.FromAddress.Valid {
		in.FromAddress = out.FromAddress
	}
	if in.ToChainId == nil || in.ToChainId.Valid() {
		in.ToChainId = (*model.BigInt)(new(big.Int).Set(utils.GetChainId(in.Chain)))
	}
	if !in.ToAddress.Valid {
		in.ToAddress = out.ToAddress
	}

	//fill empty in cross-out
	if out.FromChainId == nil || !out.FromChainId.Valid() {
		out.FromChainId = (*model.BigInt)(new(big.Int).Set((utils.GetChainId(out.Chain))))
	}
	if !out.FromAddress.Valid {
		out.FromAddress = in.FromAddress
	}
	if out.ToChainId == nil || !out.ToChainId.Valid() {
		out.ToChainId = (*model.BigInt)(new(big.Int).Set(utils.GetChainId(in.Chain)))
	}
	if !out.ToAddress.Valid {
		out.ToAddress = in.ToAddress
	}

	in.MatchHash.Scan(out.Hash)
	out.MatchHash.Scan(in.Hash)

	return nil
}
