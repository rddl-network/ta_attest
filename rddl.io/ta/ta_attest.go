package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"

	btcec "github.com/btcsuite/btcd/btcec/v2"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/gin-gonic/gin"
)

var planetmint_address string = "plmnt15xuq0yfxtd70l7jzr5hg722sxzcqqdcr8ptpl5"

func toInt(bytes []byte, offset int) int {
	result := 0
	for i := 3; i > -1; i-- {
		result = result << 8
		result += int(bytes[offset+i])
	}
	return result
}

func xorDataBlob(binary []byte, offset int, length int, is1stSegment bool, checksum byte) byte {

	//fmt.Printf("Blob Size is: %02x %d\n", length, length)
	var initializer int = 0
	if is1stSegment {
		initializer = 1
		checksum = binary[offset]
	}

	for i := initializer; i < length; i++ {
		checksum = checksum ^ binary[offset+i]
	}
	fmt.Printf("Blob Checksum: %02x\n", checksum)
	return checksum
}

func xorSegments(binary []byte) byte {
	// init variables
	size := len(binary)
	numSegments := int(binary[1])
	headersize := 8
	ext_headersize := 16
	offset := headersize + ext_headersize // that's where the data segments start

	var computed_checksum byte = byte(0)

	for i := 0; i < numSegments; i++ {
		fmt.Printf("Segment descriptior is at: %x\n", offset)
		offset += 4 // the segments load address
		length := toInt(binary, offset)
		offset += 4 // the read integer
		fmt.Printf("Blob length: %02x\n", length)
		// xor from here to offset + length for length bytes
		computed_checksum = xorDataBlob(binary, offset, length, i == 0, computed_checksum)
		offset += length
	}
	fmt.Printf(" Size : %0x, Last Data byte : %0x\n", size, offset)
	computed_checksum = computed_checksum ^ 0xEF

	return computed_checksum
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

var firmware_esp32 []byte
var firmware_esp32c3 []byte

var counter int = 0
var searchBytes []byte = []byte("RDDLRDDLRDDLRDDLRDDLRDDLRDDLRDDL")

func attestTAPublicKey(publicKey *secp256k1.PublicKey) {

	var pub_hex_string string = hex.EncodeToString(publicKey.SerializeCompressed())
	var ta string = "'{\"pubkey\": \"" + pub_hex_string + "\"}'"
	var command_str string = "/home/jeckel/go/bin/planetmint-god tx machine register-trust-anchor " + ta + " --from " + planetmint_address + " -y"
	fmt.Println("Command: " + command_str)
	cmd := exec.Command("bash", "-c", command_str)
	out, err := cmd.Output()
	if err != nil {
		// if there was any error, print it here
		fmt.Println("could not run command: ", err)
	}
	// otherwise, print the output from running the command
	fmt.Println("Output: ", string(out))
}

func computeAndSetFirmwareChecksum(patched_binary []byte) {
	binary_checksum := xorSegments(patched_binary)
	binary_size := len(patched_binary)
	patched_binary[binary_size-1] = binary_checksum
}

func getFirmware(c *gin.Context) {
	mcu := c.Param("mcu")
	privKey, pubKey := generateNewKeyPair()
	var filename string
	var fileobj []byte
	if mcu == "esp32" {
		fileobj = firmware_esp32
		filename = "tasmota32-rddl.bin"
	} else if mcu == "esp32c3" {
		fileobj = firmware_esp32c3
		filename = "tasmota32c3-rddl.bin"
	} else {
		c.String(404, "Resource not found")
		return
	}

	var patched_binary = bytes.Replace(fileobj, searchBytes, privKey.Serialize(), 1)
	computeAndSetFirmwareChecksum(patched_binary)

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, "application/octet-stream", patched_binary)

	fmt.Println(" pub key 1: ", pubKey.SerializeCompressed())
	attestTAPublicKey(pubKey)
}

func verifyBinaryIntegrity(binary []byte) bool {
	binary_size := len(binary)
	binary_checksum := xorSegments(binary)
	if binary[binary_size-1] == binary_checksum {
		fmt.Printf("The checksum is: %x\n", binary_checksum)
		return true
	} else {
		fmt.Printf("Attention: The files checksum is: %x, the computed checksum is: %x\n", binary[binary_size-1], binary_checksum)
		return false
	}
}

func generateNewKeyPair() (*secp256k1.PrivateKey, *secp256k1.PublicKey) {
	pk_source, _ := getRandomPrivateKey(32)
	privateKeyBytes, err := hex.DecodeString(pk_source)
	if err != nil {
		log.Fatalf("Failed to decode private key: %v", err)
	}

	// Initialize a secp256k1 private key object.
	privateKey, publicKey := btcec.PrivKeyFromBytes(privateKeyBytes)
	return privateKey, publicKey
}

func startWebService() {
	router := gin.Default()
	router.GET("/firmware/:mcu", getFirmware)
	router.Run("localhost:8080")
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

func loadFirmwares() {
	firmware_esp32c3 = loadFirmware("/home/jeckel/develop/rddl/ta_attest/tasmota32c3-rddl.bin")
	firmware_esp32 = loadFirmware("/home/jeckel/develop/rddl/ta_attest/tasmota32-rddl.bin")
}

func main() {
	loadFirmwares()
	startWebService()
}
