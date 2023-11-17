package config

import "sync"

const DefaultConfigTemplate = `
FIRMWARE_ESP32={{ .FirmwareESP32 }}
FIRMWARE_ESP32C3={{ .FirmwareESP32C3 }}
PLANETMINT_ACTOR={{ .PlanetmintActor }}
SERVICE_BIND={{ .ServiceBind }}
SERVICE_PORT={{ .ServicePort }}
`

// Config defines TA's top level configuration
type Config struct {
	FirmwareESP32   string `json:"firmware-esp32"    mapstructure:"firmware-esp32"`
	FirmwareESP32C3 string `json:"firmware-esp32-c3" mapstructure:"firmware-esp32-c3"`
	PlanetmintActor string `json:"planetmint-actor"  mapstructure:"planetmint-actor"`
	ServiceBind     string `json:"service-bind"      mapstructure:"service-bind"`
	ServicePort     int    `json:"service-port"      mapstructure:"service-port"`
}

// global singleton
var (
	config     *Config
	initConfig sync.Once
)

// DefaultConfig returns TA's default configuration.
func DefaultConfig() *Config {
	return &Config{
		FirmwareESP32:   "./tasmota32-rddl.bin",
		FirmwareESP32C3: "./tasmota32c3-rddl.bin",
		PlanetmintActor: "plmnt15xuq0yfxtd70l7jzr5hg722sxzcqqdcr8ptpl5",
		ServiceBind:     "localhost",
		ServicePort:     8080,
	}
}

// GetConfig returns the config instance for the SDK.
func GetConfig() *Config {
	initConfig.Do(func() {
		config = DefaultConfig()
	})
	return config
}
