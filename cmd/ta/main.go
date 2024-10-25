package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"log"
	"os"

	"github.com/planetmint/planetmint-go/app"
	"github.com/planetmint/planetmint-go/lib"
	"github.com/rddl-network/ta_attest/config"
	"github.com/rddl-network/ta_attest/service"
	"github.com/spf13/viper"
	"github.com/syndtr/goleveldb/leveldb"
)

var libConfig *lib.Config

func init() {
	encodingConfig := app.MakeEncodingConfig()
	libConfig = lib.GetConfig()
	libConfig.SetEncodingConfig(encodingConfig)
}

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
		cfg.TestnetMode = v.GetBool("TESTNET_MODE")
		cfg.DBPath = v.GetString("DB_PATH")
		cfg.PlanetmintRPCHost = v.GetString("PLANETMINT_RPC_HOST")
		cfg.LogLevel = v.GetString("LOG_LEVEL")
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

func attestFileContent(filename string, pmc service.PlanetmintClient) {
	// Open the file for reading
	file, err := os.Open(filename)
	if err != nil {
		log.Println("Error opening file:", err)
		return
	}
	defer file.Close() // Ensure file gets closed even in case of errors

	// Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)
	log.Println("Start processing the file ...")
	// Iterate over each line in the scanner
	for scanner.Scan() {
		line := scanner.Text()
		// Call your attestation function with the current line
		log.Println("Attesting : " + line)
		err := pmc.AttestTAPublicKeyHex(line)
		if err != nil {
			log.Println(err.Error())
		} else {
			log.Println("Successfully attested.")
		}
	}
	log.Println("End of file")
	// Handle any errors during scanning
	if err := scanner.Err(); err != nil {
		log.Println("Error reading file:", err)
	}
}

func main() {
	cfg, err := loadConfig("./")
	if err != nil {
		log.Fatalf("fatal error reading the configuration %s", err)
	}

	libConfig.SetChainID(cfg.PlanetmintChainID)
	grpcConn, err := service.SetupGRPCConnection(cfg)
	if err != nil {
		log.Fatalf("fatal error opening grpc connection %s", err)
	}
	pmc := service.NewPlanetmintClient(cfg.PlanetmintActor, grpcConn)

	csvFile := flag.String("attest-machine-ids-by-file", "", "Path to a new line separated machine IDs")
	flag.Parse()

	if *csvFile != "" {
		fmt.Println("Attestation mode enabled. Using CSV file:", *csvFile)

		attestFileContent(*csvFile, *pmc)
	} else {
		fmt.Println("Web Service mode")
		db, err := leveldb.OpenFile(cfg.DBPath, nil)
		if err != nil {
			log.Fatalf("fatal error opening db %s", err)
		}

		TAAttestationService := service.NewTrustAnchorAttestationService(cfg, db, pmc)
		err = TAAttestationService.Run()
		if err != nil {
			fmt.Print(err.Error())
		}
	}
}
