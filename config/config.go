package config

import "sync"

const DefaultConfigTemplate = `
FIRMWARE_ESP32="{{ .FirmwareESP32 }}"
FIRMWARE_ESP32C3="{{ .FirmwareESP32C3 }}"
PLANETMINT_ACTOR="{{ .PlanetmintActor }}"
PLANETMINT_CHAIN_ID="{{ .PlanetmintChainID }}"
SERVICE_BIND="{{ .ServiceBind }}"
SERVICE_PORT={{ .ServicePort }}"
TESTNET_MODE={{ .TestnetMode }}
DB_PATH="{{ .DBPath }}"
PLANETMINT_RPC_HOST="{{ .PlanetmintRPCHost }}"
`

// Config defines TA's top level configuration
type Config struct {
	FirmwareESP32     string `json:"firmware-esp32"      mapstructure:"firmware-esp32"`
	FirmwareESP32C3   string `json:"firmware-esp32-c3"   mapstructure:"firmware-esp32-c3"`
	PlanetmintActor   string `json:"planetmint-actor"    mapstructure:"planetmint-actor"`
	PlanetmintChainID string `json:"planetmint-chain-id" mapstructure:"planetmint-chain-id"`
	ServiceBind       string `json:"service-bind"        mapstructure:"service-bind"`
	ServicePort       int    `json:"service-port"        mapstructure:"service-port"`
	TestnetMode       bool   `json:"testnet-mode"        mapstructure:"testnet-mode"`
	DBPath            string `json:"db-path"             mapstructure:"db-path"`
	PlanetmintRPCHost string `json:"planetmint-rpc-host" mapstructure:"planetmint-rpc-host"`
}

// global singleton
var (
	config     *Config
	initConfig sync.Once
)

// DefaultConfig returns TA's default configuration.
func DefaultConfig() *Config {
	return &Config{
		FirmwareESP32:     "./tasmota32-rddl.bin",
		FirmwareESP32C3:   "./tasmota32c3-rddl.bin",
		PlanetmintActor:   "plmnt15xuq0yfxtd70l7jzr5hg722sxzcqqdcr8ptpl5",
		PlanetmintChainID: "planetmint-testnet-1",
		ServiceBind:       "localhost",
		ServicePort:       8080,
		TestnetMode:       false,
		DBPath:            "data",
		PlanetmintRPCHost: "127.0.0.1:9090",
	}
}

// GetConfig returns the config instance for the SDK.
func GetConfig() *Config {
	initConfig.Do(func() {
		config = DefaultConfig()
	})
	return config
}
