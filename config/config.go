package config

import (
	"sync"

	"github.com/rddl-network/go-utils/logger"
)

const DefaultConfigTemplate = `
PLANETMINT_ACTOR="{{ .PlanetmintActor }}"
PLANETMINT_CHAIN_ID="{{ .PlanetmintChainID }}"
SERVICE_BIND="{{ .ServiceBind }}"
SERVICE_PORT={{ .ServicePort }}
TESTNET_MODE={{ .TestnetMode }}
DB_PATH="{{ .DBPath }}"
PLANETMINT_RPC_HOST="{{ .PlanetmintRPCHost }}"
LOG_LEVEL="{{ .LogLevel }}"
`

// Config defines TA's top level configuration
type Config struct {
	PlanetmintActor   string `json:"planetmint-actor"    mapstructure:"planetmint-actor"`
	PlanetmintChainID string `json:"planetmint-chain-id" mapstructure:"planetmint-chain-id"`
	ServiceBind       string `json:"service-bind"        mapstructure:"service-bind"`
	ServicePort       int    `json:"service-port"        mapstructure:"service-port"`
	TestnetMode       bool   `json:"testnet-mode"        mapstructure:"testnet-mode"`
	DBPath            string `json:"db-path"             mapstructure:"db-path"`
	PlanetmintRPCHost string `json:"planetmint-rpc-host" mapstructure:"planetmint-rpc-host"`
	LogLevel          string `json:"log-level"           mapstructure:"log-level"`
}

// global singleton
var (
	config     *Config
	initConfig sync.Once
)

// DefaultConfig returns TA's default configuration.
func DefaultConfig() *Config {
	return &Config{
		PlanetmintActor:   "plmnt15xuq0yfxtd70l7jzr5hg722sxzcqqdcr8ptpl5",
		PlanetmintChainID: "planetmint-testnet-1",
		ServiceBind:       "localhost",
		ServicePort:       8080,
		TestnetMode:       false,
		DBPath:            "data",
		PlanetmintRPCHost: "127.0.0.1:9090",
		LogLevel:          logger.DEBUG,
	}
}

// GetConfig returns the config instance for the SDK.
func GetConfig() *Config {
	initConfig.Do(func() {
		config = DefaultConfig()
	})
	return config
}
