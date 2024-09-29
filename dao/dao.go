package dao

import (
	"app/model"
	"app/utils"
	"database/sql"
	"fmt"
	"github.com/ethereum/go-ethereum/log"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/zeromicro/go-zero/core/stores/builder"
	"github.com/zeromicro/go-zero/core/stringx"
	"strings"
	"sync"
	"time"
)

const (
	SQL_MAX_PLACEHOLDERS = 65535
	MaxIdleConns         = 10
	MaxOpenConns         = 100
)

const dupErr = "pq: duplicate key value violates unique constraint"

var (
	resultInsertFieldNames = builder.RawFieldNames(&model.Result{}, true)
	resultInsertRows       = strings.Join(stringx.Remove(resultInsertFieldNames, "id"), ",")
	resultInsertTags       = strings.Join(slicesWithPrefix(stringx.Remove(resultInsertFieldNames, "id"), ":"), ",")

	resultUpdateFieldNames = []string{"match_id", "from_chain_id", "from_address", "to_chain_id", "to_address", "real_token_out", "real_amount_out", "real_token_in", "real_amount_in"}
	resultUpdateRows       = strings.Join(resultUpdateFieldNames, ",")
	resultUpdateTags       = builder.PostgreSqlJoin(resultUpdateFieldNames)

	MatchedResultUpdateFieldNames = []string{"match_id", "match_hash", "from_chain_id", "to_chain_id", "from_address", "to_address"}
	MatchedResultUpdateTags       = builder.PostgreSqlJoin(MatchedResultUpdateFieldNames)

	FillerResultUpdateFieldNames = []string{"real_token_out", "real_amount_out", "real_token_in", "real_amount_in"}
	FillerResultUpdateTags       = builder.PostgreSqlJoin(FillerResultUpdateFieldNames)

	UpdateTxFromAddress     = []string{"tx_from_address"}
	UpdateTxFromAddressTags = builder.PostgreSqlJoin(UpdateTxFromAddress)
)

type Dao struct {
	db    *sqlx.DB
	table string
}

func NewDao(host string) *Dao {
	db, err := sqlx.Connect("postgres", host)
	if err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(MaxOpenConns)
	db.SetMaxIdleConns(MaxIdleConns)
	return &Dao{
		db:    db,
		table: "cross_chain",
	}
}

func (d *Dao) LatestId() (latest uint64, err error) {
	stmt := fmt.Sprintf("select max(id) from %s", d.table)
	err = d.db.Get(&latest, stmt)
	return
}

func (d *Dao) MinUnmatchId(project string) (min_id uint64, err error) {
	stmt := fmt.Sprintf("select min(id) from %s where direction = '%s' and project = '%s' and match_id is null", d.table, model.InDirection, project)
	err = d.db.Get(&min_id, stmt)
	return
}

func (d *Dao) Save(results model.Results) (err error) {
	tx, err := d.db.Beginx()
	if err != nil {
		return
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	maxInsert := SQL_MAX_PLACEHOLDERS / 30
	for i := 0; i < len(results); i += maxInsert {
		batch := results[i:utils.Min(i+maxInsert, len(results))]
		err = d.save(tx, batch)
		if err != nil {
			return
		}
	}
	return nil
}

func (d *Dao) save(tx *sqlx.Tx, results model.Results) error {
	if len(results) == 0 {
		return nil
	}
	for _, r := range results {
		if len(r.Detail) == 0 {
			r.Detail = []byte(`{}`)
		}
	}
	stmt := fmt.Sprintf("insert into %s (%s) values (%s)", d.table, resultInsertRows, resultInsertTags)
	_, err := tx.NamedExec(stmt, results)
	return err
}

func (d *Dao) SaveNew(results model.Results) {
	var size = len(results)
	if len(results) > 3000 {
		size = len(results) / 10
	}
	var wg = &sync.WaitGroup{}

	for i := 0; i <= len(results); i += size {
		wg.Add(1)
		go func(i int, data model.Results) {
			right := utils.Min(i+size, len(data))
			if right < i {
				i = len(data) - 1
				right = len(data)
			}
			d.saveNew(data[i:right])
			wg.Done()
		}(i, results)
	}
	wg.Wait()
}

func (d *Dao) GetUnmatches(chain, name string, from, to time.Time) model.Results {
	var ins model.Results

	chainId := utils.GetChainId(chain)
	stmt := fmt.Sprintf("select * from %s where direction = 'in' and match_id is null and project = '%s' and ts <= '%s' and from_chain_id = %d order by ts asc", d.table, name, to.String()[:19], chainId.Int64())
	err := d.db.Select(&ins, stmt)
	if err != nil {
		log.Error("get unmatched data failed", "Error", err)
		return nil
	}
	return ins
}

func (d *Dao) saveNew(results model.Results) {
	for _, result := range results {
		if len(result.Detail) == 0 {
			result.Detail =
				[]byte(`{}`)
		}
		stmt := fmt.Sprintf("insert into %s (%s) values (%s)", d.table, resultInsertRows, resultInsertTags)
		_, err := d.db.NamedExec(stmt, result)
		if err == nil || strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint") {
			continue
		} else {
			log.Error("save single failed", "ERR", err, "Hash", result.Hash, "FromChainId", result.FromChainId, "ToChainId", result.ToChainId)
		}
	}
}

func (d *Dao) UpdateMatchResult(results model.Results) (err error) {
	if len(results) == 0 {
		return nil
	}
	tx, err := d.db.Beginx()
	if err != nil {
		return
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			log.Error("save trades rollback in UpdateMatchResult")
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()
	stmt := fmt.Sprintf("update %s set %s where id = $1", d.table, MatchedResultUpdateTags)
	for _, r := range results {
		_, err = d.db.Exec(stmt, r.Id, r.MatchId, r.MatchHash, r.FromChainId.String(), r.ToChainId.String(), r.FromAddress, r.ToAddress)
	}
	log.Info("update matched results done", "number", len(results))
	return
}

func (d *Dao) LastUpdate(chain, project string) (uint64, error) {
	var last uint64
	stmt := fmt.Sprintf("select number from %s where chain = $1 and project = $2 order by number desc limit 1", d.table)
	err := d.db.Get(&last, stmt, chain, project)
	if err == sql.ErrNoRows {
		err = nil
	}
	return last, err
}

func (d *Dao) DB() *sqlx.DB { return d.db }

func (d *Dao) Table() string { return d.table }

func slicesWithPrefix(s []string, prefix string) []string {
	ret := make([]string, len(s))
	for i := range s {
		ret[i] = prefix + s[i]
	}
	return ret
}

func getSqlxNamedTagsForUpdate(s []string) []string {
	ret := make([]string, len(s))
	for i := range s {
		ret[i] = fmt.Sprintf(`%s=:%s`, s[i], s[i])
	}
	return ret
}
