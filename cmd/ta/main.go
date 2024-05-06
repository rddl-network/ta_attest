package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"os"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/rddl-network/ta_attest/config"
	"github.com/rddl-network/ta_attest/service"
	"github.com/syndtr/goleveldb/leveldb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/spf13/viper"
)

func loadConfig(path string) (cfg *config.Config, err error) {
	v := viper.New()
	v.AddConfigPath(path)
	v.SetConfigName("app")
	v.SetConfigType("env")
	v.AutomaticEnv()

	err = v.ReadInConfig()
	if err == nil {
		cfg = config.GetConfig()
		cfg.ServiceBind = v.GetString("SERVICE_BIND")
		cfg.ServicePort = v.GetInt("SERVICE_PORT")
		cfg.PlanetmintActor = v.GetString("PLANETMINT_ACTOR")
		cfg.PlanetmintChainID = v.GetString("PLANETMINT_CHAIN_ID")
		cfg.FirmwareESP32 = v.GetString("FIRMWARE_ESP32")
		cfg.FirmwareESP32C3 = v.GetString("FIRMWARE_ESP32C3")
		cfg.TestnetMode = v.GetBool("TESTNET_MODE")
		cfg.DBPath = v.GetString("DB_PATH")
		cfg.PlanetmintRPCHost = v.GetString("PLANETMINT_RPC_HOST")
		return
	}
	log.Println("no config file found")

	tmpl := template.New("appConfigFileTemplate")
	configTemplate, err := tmpl.Parse(config.DefaultConfigTemplate)
	if err != nil {
		return
	}

	var buffer bytes.Buffer
	if err = configTemplate.Execute(&buffer, config.GetConfig()); err != nil {
		return
	}

	if err = v.ReadConfig(&buffer); err != nil {
		return
	}
	if err = v.SafeWriteConfig(); err != nil {
		return
	}

	log.Println("default config file created. please adapt it and restart the application. exiting...")
	os.Exit(0)
	return
}

func main() {
	cfg, err := loadConfig("./")
	if err != nil {
		log.Fatalf("fatal error reading the configuration %s", err)
	}

	db, err := leveldb.OpenFile(cfg.DBPath, nil)
	if err != nil {
		log.Fatalf("fatal error opening db %s", err)
	}

	grpcConn, err := grpc.Dial(
		cfg.PlanetmintRPCHost,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.ForceCodec(codec.NewProtoCodec(nil).GRPCCodec())),
	)
	if err != nil {
		log.Fatalf("fatal error opening grpc connection %s", err)
	}
	pmc := service.NewPlanetmintClient(cfg.PlanetmintActor, grpcConn)

	TAAttestationService := service.NewTrustAnchorAttestationService(cfg, db, pmc)
	err = TAAttestationService.Run()
	if err != nil {
		fmt.Print(err.Error())
	}
}
