package aggregator

import (
	crosschain "app/cross_chain"
	"app/model"
	"app/provider"
	"app/svc"
	"app/utils"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/log"
)

var Web3QueryStartBlock = map[string]uint64{
	"ethereum":  16028978,
	"eth":       16028978,
	"bsc":       21051775,
	"polygon":   31713584,
	"fantom":    46370676,
	"arbitrum":  82939057,
	"avalanche": 19548760,
	"optimism":  14299344,
}

var (
	BatchSize uint64
)

type Aggregator struct {
	svc        *svc.ServiceContext
	chain      string
	provider   *provider.Provider
	collectors []model.Collector
}

func NewAggregator(svc *svc.ServiceContext, chain string) *Aggregator {
	p := svc.Providers.Get(chain)
	if p == nil {
		panic(fmt.Sprintf("%v: invalid provider", chain))
	}
	return &Aggregator{
		svc:        svc,
		chain:      chain,
		provider:   p,
		collectors: crosschain.GetCollectors(svc, chain),
	}
}

func (a *Aggregator) Start() {
	for _, c := range a.collectors {
		if a.chain == "arbitrum-nova" {
			continue
		}
		go a.DoJob(c)
		time.Sleep(time.Second)
	}
}

func (a *Aggregator) DoJob(c model.Collector) {
	a.svc.Wg.Add(1)
	defer a.svc.Wg.Done()

	if c.Name() == "WormHole" {
		time.Sleep(time.Duration(rand.Intn(600)) * time.Second)
	}

	timer := time.NewTimer(1 * time.Second)
	last := Web3QueryStartBlock[a.chain]
	batchSize := BatchSize
	var lasterr error
	for {
		select {
		case <-a.svc.Ctx.Done():
			return
		case <-timer.C:
			latest, err := a.provider.LatestNumber()
			if err != nil {
				log.Error("get latest number failed", "chain", a.chain, "err", err.Error())
				break
			}
			cnt429 := 0
			if latest < last {
				log.Error("latest < last", "chain", a.chain, "latest", latest, "last", last)
			}
			for last < latest {
				var shouldBreak bool
				select {
				case <-a.svc.Ctx.Done():
					shouldBreak = true
				default:
				}
				if shouldBreak {
					break
				}
				right := utils.Min(latest, last+batchSize)
				fetched, err := a.Work(c, last+1, right)
				if err == utils.ErrTooManyRecords || err == utils.ErrGethRespTooLarge {
					batchSize = batchSize / 2
					log.Warn("too many req records", "chain", a.chain, "project", c.Name(), "batch size", batchSize)
				} else if err == utils.ErrNetwork429 {
					cnt429++
					time.Sleep(time.Duration(cnt429*30) * time.Second)
					if cnt429%6 == 0 {
						log.Error("network status 429", "chain", a.chain, "project", c.Name(), "cnt429", cnt429, "from", last+1, "to", right, "err", err)
					}
				} else if err != nil {
					if err == utils.ErrEtherscanRateLimit {
						log.Warn("etherscan rate limit", "chain", a.chain, "project", c.Name())
					} else if lasterr != nil && lasterr.Error() != err.Error() {
						lasterr = err
						log.Error("job failed", "chain", a.chain, "project", c.Name(), "from", last+1, "to", right, "err", err)
					}
				} else {
					cnt429 = 0
					last = right
					log.Info("collect done", "chain", a.chain, "project", c.Name(), "current number", last, "batch size", batchSize)
					if fetched < utils.EtherScanMaxResult*0.3 && batchSize <= 3*utils.EtherScanMaxResult {
						batchSize += 100
					}
				}
			}
		}
		timer.Reset(180 * time.Second)
	}
}

func (a *Aggregator) Work(c model.Collector, from, to uint64) (int, error) {
	var totalFetched int
	var results model.Results
	//log.Info("Aggregator begins", "collector", c.Name(), "from", from)
	switch v := c.(type) {
	case model.EventCollector:
		topics0 := v.Topics0(a.chain)
		events, err := a.provider.GetLogs(topics0, from, to)
		if err != nil {
			return 0, err
		}
		totalFetched = len(events)
		sort.Sort(events)
		results = v.Extract(a.chain, events)
	case model.MsgCollector:
		addrs := c.(model.MsgCollector).Contracts(a.chain)
		if len(addrs) == 0 {
			return 0, nil
		}
		selectors := v.Selectors(a.chain)
		calls, err := a.provider.GetCalls(addrs, selectors, from, to)
		if err != nil {
			return 0, err
		}
		totalFetched = len(calls)
		results = v.Extract(a.chain, calls)
	case model.TransferCollector:
		addrs := c.(model.TransferCollector).Addresses(a.chain)
		if len(addrs) == 0 {
			return 0, nil
		}
		msgs, err := a.provider.GetERC20Transfer(addrs, from, to)
		if err != nil {
			return 0, err
		}
		totalFetched = len(msgs)
		sort.Sort(msgs)
		results = v.Extract(a.chain, msgs)
	default:
		panic("invalid collector")
	}
	//res := filterDuplicates(results)
	err := a.svc.Dao.Save(results)
	if err != nil {
		if !strings.Contains(err.Error(), "duplicate key") {
			log.Error("result save failed and then save by single", "chain", a.chain, "project", c.Name(), "from", from, "to", to, "err", err)
		}
		a.svc.Dao.SaveNew(results)
	}
	return totalFetched, nil
}

func (a *Aggregator) getAddressesFirstInvocation(addresses []string) (uint64, error) {
	nums := make([]uint64, 0)
	for _, addr := range addresses {
		n, err := a.provider.GetContractFirstInvocation(addr)
		if err != nil {
			log.Error("get address first invoke failed", "chain", a.chain, "address", addr, "err", err.Error())
		}
		if n != 0 {
			nums = append(nums, n)
		}
	}
	if len(nums) == 0 {
		return 0, nil
	}
	return utils.Min(nums...), nil
}

func (a *Aggregator) getCkpt(project string) (uint64, error) {
	last, err := a.svc.Dao.LastUpdate(a.chain, project)
	if err != nil {
		return 0, err
	}
	if last == 0 {
		last = Web3QueryStartBlock[a.chain]
	}
	if last == 0 {
		last, err = a.provider.LatestNumber()
		if err != nil {
			return 0, err
		} else {
			last = last - 1000000
		}
	}
	return last, nil
}

func (a *Aggregator) GetDeployer(addresses []string) ([]string, error) {
	deployers := make([]string, 0)
	for _, addr := range addresses {
		deployer, err := a.provider.GetContractDeployer(addr)
		if err != nil {
			log.Error("get address first invoke failed", "chain", a.chain, "address", addr, "err", err.Error())
		}
		if len(deployer) != 0 {
			deployers = append(deployers, deployer)
		}
	}
	if len(deployers) == 0 {
		return nil, nil
	}
	return deployers, nil
}
