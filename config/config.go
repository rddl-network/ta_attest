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
	FirmwareESP32   string `mapstructure:"firmware-esp32" json:"firmware-esp32"`
	FirmwareESP32C3 string `mapstructure:"firmware-esp32c3" json:"firmware-esp32c3"`
	PlanetmintActor string `mapstructure:"planetmint-actor" json:"planetmint-actor"`
	ServiceBind     string `mapstructure:"service-bind" json:"service-bind"`
	ServicePort     int    `mapstructure:"service-port" json:"service-port"`
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
