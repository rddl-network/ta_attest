package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	btcec "github.com/btcsuite/btcd/btcec/v2"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/gin-gonic/gin"
	"github.com/rddl-network/ta_attest/config"
	"github.com/spf13/viper"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/planetmint/planetmint-go/app"
	"github.com/planetmint/planetmint-go/lib"
	machinetypes "github.com/planetmint/planetmint-go/x/machine/types"
)

func LoadConfig(path string) (v *viper.Viper, err error) {
	v = viper.New()
	v.AddConfigPath(path)
	v.SetConfigName("app")
	v.SetConfigType("env")

	v.AutomaticEnv()

	err = v.ReadInConfig()
	if err == nil {
		return
	}
	log.Println("no config file found.")

	tmpl := template.New("appConfigFileTemplate")
	configTemplate, err := tmpl.Parse(config.DefaultConfigTemplate)
	if err != nil {
		return
	}

	var buffer bytes.Buffer
	err = configTemplate.Execute(&buffer, config.GetConfig())
	if err != nil {
		return
	}

	err = v.ReadConfig(&buffer)
	if err != nil {
		return
	}
	err = v.SafeWriteConfig()
	if err != nil {
		return
	}
	log.Println("default config file created. please adapt it and restart the application. exiting...")
	os.Exit(0)
	return
}

var planetmintAddress string
var planetmintGo string
var libConfig *lib.Config

func init() {
	encodingConfig := app.MakeEncodingConfig()

	libConfig = lib.GetConfig()
	libConfig.SetEncodingConfig(encodingConfig)
}

func toInt(bytes []byte, offset int) int {
	result := 0
	for i := 3; i > -1; i-- {
		result = result << 8
		result += int(bytes[offset+i])
	}
	return result
}

func xorDataBlob(binary []byte, offset int, length int, is1stSegment bool, checksum byte) byte {

	initializer := 0
	if is1stSegment {
		initializer = 1
		checksum = binary[offset]
	}

	for i := initializer; i < length; i++ {
		checksum = checksum ^ binary[offset+i]
	}
	return checksum
}

func xorSegments(binary []byte) byte {
	// init variables
	numSegments := int(binary[1])
	headersize := 8
	extHeadersize := 16
	offset := headersize + extHeadersize // that's where the data segments start

	computedChecksum := byte(0)

	for i := 0; i < numSegments; i++ {
		offset += 4 // the segments load address
		length := toInt(binary, offset)
		offset += 4 // the read integer
		// xor from here to offset + length for length bytes
		computedChecksum = xorDataBlob(binary, offset, length, i == 0, computedChecksum)
		offset += length
	}
	computedChecksum = computedChecksum ^ 0xEF

	return computedChecksum
}

func randomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func getRandomPrivateKey(n int) (string, error) {
	return randomHex(n)
}

var firmwareESP32 []byte
var firmwareESP32C3 []byte
var searchBytes = []byte("RDDLRDDLRDDLRDDLRDDLRDDLRDDLRDDL")

func attestTAPublicKeyHex(pubHexString string) error {
	addr := sdk.MustAccAddressFromBech32(planetmintAddress)
	msg := machinetypes.NewMsgRegisterTrustAnchor(planetmintAddress, &machinetypes.TrustAnchor{
		Pubkey: pubHexString,
	})

	ctx := context.Background()
	txBytes, _, err := lib.BuildAndSignTx(ctx, addr, msg)
	if err != nil {
		return err
	}

	_, err = lib.BroadcastTx(ctx, txBytes)
	if err != nil {
		return err
	}
	return nil
}

func attestTAPublicKey(publicKey *secp256k1.PublicKey) error {
	pubHexString := hex.EncodeToString(publicKey.SerializeCompressed())
	return attestTAPublicKeyHex(pubHexString)
}

func postPubKey(c *gin.Context) {
	pubkey := c.Param("pubkey")
	_, err := hex.DecodeString(pubkey)
	if err == nil {
		err = attestTAPublicKeyHex(pubkey)
		if err == nil {
			c.IndentedJSON(http.StatusOK, pubkey)
		} else {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		}
	} else {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid pubkey"})
	}
}

func computeAndSetFirmwareChecksum(patchedBinary []byte) {
	binaryChecksum := xorSegments(patchedBinary)
	binarySize := len(patchedBinary)
	patchedBinary[binarySize-1] = binaryChecksum
}

func getFirmware(c *gin.Context) {
	mcu := c.Param("mcu")
	privKey, pubKey := generateNewKeyPair()
	var filename string
	var fileobj []byte
	if mcu == "esp32" {
		fileobj = firmwareESP32
		filename = "tasmota32-rddl.bin"
	} else if mcu == "esp32c3" {
		fileobj = firmwareESP32C3
		filename = "tasmota32c3-rddl.bin"
	} else {
		c.String(404, "Resource not found, Firmware not supported")
		return
	}

	var patchedBinary = bytes.Replace(fileobj, searchBytes, privKey.Serialize(), 1)
	computeAndSetFirmwareChecksum(patchedBinary)

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, "application/octet-stream", patchedBinary)

	fmt.Println(" pub key 1: ", pubKey.SerializeCompressed())
	_ = attestTAPublicKey(pubKey)
}

func verifyBinaryIntegrity(binary []byte) bool {
	binarySize := len(binary)
	binaryChecksum := xorSegments(binary)
	if binary[binarySize-1] == binaryChecksum {
		fmt.Printf("The integrity of the file got verified. The checksum is: %x\n", binaryChecksum)
		return true
	}
	fmt.Printf("Attention: File integrity check FAILED. The files checksum is: %x, the computed checksum is: %x\n", binary[binarySize-1], binaryChecksum)
	return false
}

func generateNewKeyPair() (*secp256k1.PrivateKey, *secp256k1.PublicKey) {
	pkSource, _ := getRandomPrivateKey(32)
	privateKeyBytes, err := hex.DecodeString(pkSource)
	if err != nil {
		log.Fatalf("Failed to decode private key: %v", err)
	}

	// Initialize a secp256k1 private key object.
	privateKey, publicKey := btcec.PrivKeyFromBytes(privateKeyBytes)
	return privateKey, publicKey
}

func startWebService(config *viper.Viper) error {
	router := gin.Default()
	router.GET("/firmware/:mcu", getFirmware)
	router.POST("/register/:pubkey", postPubKey)

	bindAddress := config.GetString("SERVICE_BIND")
	servicePort := config.GetString("SERVICE_PORT")
	err := router.Run(bindAddress + ":" + servicePort)
	return err
}

func loadFirmware(filename string) []byte {
	content, err := os.ReadFile(filename)
	if err != nil {
		panic("could not read firmware")
	}

	if !verifyBinaryIntegrity(content) {
		panic("given firmware integrity check failed")
	}

	return content
}

func loadFirmwares(config *viper.Viper) {
	esp32 := config.GetString("FIRMWARE_ESP32")
	esp32c3 := config.GetString("FIRMWARE_ESP32C3")

	firmwareESP32 = loadFirmware(esp32)
	firmwareESP32C3 = loadFirmware(esp32c3)
}

func main() {
	config, err := LoadConfig("./")
	if err != nil {
		log.Fatalf("fatal error config file: %s", err)
	}

	planetmintGo = config.GetString("PLANETMINT_GO")
	planetmintAddress = config.GetString("PLANETMINT_ACTOR")
	if err != nil || planetmintAddress == "" || planetmintGo == "" {
		panic("couldn't read configuration")
	}
	fmt.Printf("global config %s\n", planetmintAddress)
	loadFirmwares(config)
	err = startWebService(config)
	if err != nil {
		fmt.Print(err.Error())
	}

}
