package crosschain

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
	"app/model"
	"app/svc"
)

func GetCollectors(svc *svc.ServiceContext, chain string) []model.Collector {
	arbi := arbitrumBridge.NewArbiBridge()
	opt := optimismGateway.NewOptiCollector()
	avax := avaxBridge.NewAvaxTransferCollector()
	across := across.NewAcrossCollector()
	synapse := synapse.NewSynapseCollector(svc)
	hop := hop.NewHopCollector()
	stargate := stargate.NewStargateCollector(svc)
	anyswap := anyswap.NewAnyswapCollector(svc)
	//ren := renbridge.NewRenbridgeCollector(),
	worm := wormhole.NewWormHoleCollector(svc)
	cbridge := celer_bridge.NewCBridgeCollector()
	poly := poly.NewPolyCollector()

	var chainMapCollectors = map[string][]model.Collector{
		"optimism":  {opt, cbridge, across, anyswap, worm, poly, stargate, hop},
		"eth":       {arbi, opt, avax, stargate, cbridge, across, worm, poly, anyswap, stargate, synapse, hop},
		"bsc":       {stargate, anyswap, cbridge, across, worm, poly, hop},
		"polygon":   {synapse, stargate, anyswap, cbridge, across, worm, poly, hop},
		"fantom":    {stargate, across, anyswap, synapse, poly, hop},
		"arbitrum":  {synapse, stargate, anyswap, cbridge, across, worm, poly, hop},
		"avalanche": {synapse, stargate, anyswap, cbridge, across, worm, poly, hop},
	}
	return chainMapCollectors[chain]
}
