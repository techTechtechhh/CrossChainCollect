package main

import (
	"app/aggregator"
	"app/cmd"
	"app/config"
	"app/provider/chainbase"
	"app/provider/etherscan"
	"app/svc"
	"context"
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"os"
	"os/signal"
	"syscall"

	"net/http"
	_ "net/http/pprof"

	"github.com/ethereum/go-ethereum/log"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "./config.yaml", "the config file")

func main() {
	flag.Parse()
	var cfg config.Config
	config.LoadCfg(&cfg, *configFile)
	conf.MustLoad(*configFile, &cfg.Rest)

	logx.DisableStat()
	server := rest.MustNewServer(cfg.Rest)
	defer server.Stop()

	lvl, err := log.LvlFromString(cfg.LogLvl)
	if err != nil {
		panic(err)
	}
	aggregator.BatchSize = cfg.BatchSize
	go func() {
		if cfg.Pprof != 0 {
			fmt.Println(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%v", cfg.Pprof), nil))
		}
	}()

	go func() {
		prometheus.MustRegister(etherscan.CrossChainRequestCount)
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":2112", nil)
	}()

	fmt.Println("log level:", lvl.String(), "\nbatch size:", cfg.BatchSize, "\npprof port:", cfg.Pprof, "\nchainbase rate:", cfg.ChainbaseLimit)

	log.Root().SetHandler(log.MultiHandler(
		log.LvlFilterHandler(lvl, log.StreamHandler(os.Stderr, log.TerminalFormat(true))),
	))
	ctx, cancel := context.WithCancel(context.Background())
	chainbase.SetupLimit(cfg.ChainbaseLimit)
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
		<-sig
		cancel()
	}()

	srvCtx := svc.NewServiceContext(ctx, &cfg)
	cmd.RegisterHandlers(server, srvCtx)
	fmt.Printf("Starting server at %s:%d...\n", cfg.Rest.Host, cfg.Rest.Port)
	go server.Start()
	for name := range srvCtx.Config.ChainProviders {
		agg := aggregator.NewAggregator(srvCtx, name)
		go agg.Start()
	}
	<-ctx.Done()
	srvCtx.Wg.Wait()
	fmt.Println("exit")
}
