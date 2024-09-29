package main

import (
	"app/cmd"
	"app/config"
	"app/matcher"
	"app/svc"
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ethereum/go-ethereum/log"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "./config.yaml", "the config file")
var port = flag.Int("p", 7001, "listening port")

func main() {
	flag.Parse()
	var cfg config.Config
	config.LoadCfg(&cfg, *configFile)
	cfg.Rest.Port = *port
	conf.MustLoad(*configFile, &cfg.Rest)

	logx.DisableStat()
	server := rest.MustNewServer(cfg.Rest)
	defer server.Stop()
	go func() {
		if cfg.Pprof != 0 {
			fmt.Println(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%v", cfg.Pprof), nil))
		}
	}()
	log.Root().SetHandler(log.MultiHandler(
		log.LvlFilterHandler(log.LvlInfo, log.StreamHandler(os.Stderr, log.TerminalFormat(true))),
	))
	ctx, cancel := context.WithCancel(context.Background())
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
	m := matcher.NewMatcher(srvCtx)
	m.Start()
	<-ctx.Done()
	srvCtx.Wg.Wait()
	fmt.Println("exit")
}
