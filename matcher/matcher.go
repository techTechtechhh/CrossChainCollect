package matcher

import (
	"app/cross_chain/across"
	"app/cross_chain/anyswap"
	"app/cross_chain/arbitrumBridge"
	"app/cross_chain/avaxBridge"
	"app/cross_chain/celer_bridge"
	"app/cross_chain/hop"
	"app/cross_chain/optimismGateway"
	"app/cross_chain/poly"
	"app/cross_chain/stargate"
	"app/cross_chain/synapse"
	"app/cross_chain/wormhole"
	"app/dao"
	"app/model"
	"app/svc"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/log"
	_ "github.com/lib/pq"
)

const (
	interval  = 5 * 60
	batchSize = 1000
	ChanLimit = 20
)

type Matcher struct {
	svc      *svc.ServiceContext
	projects map[string]model.Matcher
}

func NewMatcher(svc *svc.ServiceContext) *Matcher {
	return &Matcher{
		svc: svc,
		projects: map[string]model.Matcher{
			synapse.NewSynapseCollector(svc).Name():   NewSimpleInMatcher(svc),
			celer_bridge.NewCBridgeCollector().Name(): NewSimpleInMatcher(svc),
			stargate.NewStargateCollector(nil).Name(): NewSimpleInMatcher(svc),
			anyswap.NewAnyswapCollector(nil).Name():   NewSimpleInMatcher(svc),
			across.NewAcrossCollector().Name():        NewSimpleInMatcher(svc),
			poly.NewPolyCollector().Name():            NewSimpleInMatcher(svc),
			arbitrumBridge.NewArbiBridge().Name():     NewSimpleInMatcher(svc),
			poly.NewPolyCollector().Name():            NewSimpleInMatcher(svc),
			celer_bridge.NewCBridgeCollector().Name(): NewSimpleInMatcher(svc),
			anyswap.NewAnyswapCollector(nil).Name():   NewSimpleInMatcher(svc),
			across.NewAcrossCollector().Name():        NewSimpleInMatcher(svc),
			wormhole.NewWormHoleCollector(nil).Name(): NewSimpleInMatcher(svc),
			hop.NewHopCollector().Name():              NewSimpleInMatcher(svc),
			arbitrumBridge.NewArbiBridge().Name():     NewSimpleInMatcher(svc),
			optimismGateway.NewOptiCollector().Name(): NewSimpleInMatcher(svc),
			avaxBridge.NewAvaxEventCollector().Name(): NewSimpleInMatcher(svc),
		},
	}
}

func (m *Matcher) Start() {
	for project, matcher := range m.projects {
		m.svc.Wg.Add(1)
		go m.StartMatch(project, matcher)
	}
}

func (m *Matcher) StartMatch(project string, matcher model.Matcher) {
	defer m.svc.Wg.Done()
	log.Info("matcher start", "project", project)
	timer := time.NewTimer(1 * time.Second)
	latest, err := m.svc.Dao.LatestId()
	min, err := m.svc.Dao.MinUnmatchId(project)
	left := min + 1
	latest++

	for {
		select {
		case <-m.svc.Ctx.Done():
			return
		case <-timer.C:
			if err != nil {
				log.Error("get latst id failed", "project", project, "err", err)
				break
			}
			for min < left {
				var shouldBreak bool
				select {
				case <-m.svc.Ctx.Done():
					shouldBreak = true
				default:
				}
				if shouldBreak {
					break
				}
				matched, err, left := m.BeginMatch(min, latest, project, matcher)
				if err != nil {
					log.Error("match job failed", "project", project, "from", latest+1, "to", left, "err", err)
				} else if left <= min {
					log.Info("match over", "project", project, "left", left, "latest", latest, "total matched", matched)
					return
				} else {
					latest = left
					if matched > 0 {
						log.Info("match done", "project", project, "left", left, "latest", latest, "total matched", matched)
					}
				}
				time.Sleep(5 * time.Second)
			}
		}
		timer.Reset(interval * time.Second)
	}
}

func (m *Matcher) BeginMatch(from, to uint64, project string, matcher model.Matcher) (matched int, err error, left uint64) {
	var stmt string
	switch matcher.(type) {
	case *SimpleInMatcher:
		stmt = fmt.Sprintf("select * from %s where id <= $1 and match_id is null and project = '%s' and direction = '%s' and ts <= '2024-01-02' and match_tag not in ('0', '1', '2', '3', '4') order by id desc limit 5000 ", m.svc.Dao.Table(), project, model.InDirection)
	default:
		panic("invalid matcher")
	}
	var results model.Results
	err = m.svc.Dao.DB().Select(&results, stmt, to)
	if err != nil {
		return
	}
	step := 500
	batches := len(results) / step
	matchChan := make(chan int)
	wg2 := sync.WaitGroup{}
	for i := 0; i < batches; i++ {
		startIndex := i * step
		endIndex := (i + 1) * step
		if endIndex > len(results) {
			endIndex = len(results)
		}
		batch := results[startIndex:endIndex]
		wg2.Add(1)

		go match(batch, m.svc.Dao, matcher, matchChan, &wg2)
	}
	go func() {
		wg2.Wait()
		close(matchChan)
	}()
	if len(results) == 0 {
		left = from - 1
	} else {
		left = results[len(results)-1].Id
	}
	for result := range matchChan {
		matched += result
	}
	return
}

func match(results model.Results, dao *dao.Dao, matcher model.Matcher, matched chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	shouldUpdates, _, _, err := matcher.Match(results)
	if err != nil {
		return
	}
	err = dao.UpdateMatchResult(shouldUpdates)
	if err != nil {
		return
	}
	match := len(shouldUpdates)
	matched <- match
}
