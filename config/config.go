package config

import (
	"io/ioutil"

	"github.com/zeromicro/go-zero/rest"
	"gopkg.in/yaml.v3"
)

type Database struct {
	CrosschainDataSource string `yaml:"CrosschainDataSource"`
}

type ChainProvider struct {
	Node            string   `yaml:"Node"`
	ScanUrl         string   `yaml:"ScanUrl"`
	ApiKeys         []string `yaml:"ApiKeys"`
	ChainbaseTable  string   `yaml:"ChainbaseTable"`
	EnableTraceCall bool     `yaml:"EnableTraceCall"`
}

type Config struct {
	Rest               rest.RestConf
	LogLvl             string                    `yaml:"LogLvl"`
	Database           Database                  `yaml:"Database"`
	Proxy              []string                  `yaml:"Proxy"`
	BatchSize          uint64                    `yaml:"BatchSize"`
	Pprof              int                       `yaml:"Pprof"`
	EtherscanRateLimit int                       `yaml:"EtherscanRateLimit"`
	ChainbaseApiKey    string                    `yaml:"ChainbaseApiKey"`
	ChainbaseLimit     int                       `yaml:"ChainbaseLimit"`
	ChainProviders     map[string]*ChainProvider `yaml:"ChainProviders"`
}

func LoadCfg[T any](v T, path string) {
	f, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(f, v)
	if err != nil {
		panic(err)
	}
}

var ChainName = map[string]string{
	"1":     "eth",
	"10":    "optimism",
	"56":    "bsc",
	"137":   "polygon",
	"250":   "fantom",
	"42161": "arbitrum",
	"43114": "avalanche",
}
