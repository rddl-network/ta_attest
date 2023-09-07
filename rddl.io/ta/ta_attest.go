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

var reference_content []byte
var counter int = 0
var searchBytes []byte = []byte("RDDLRDDLRDDLRDDLRDDLRDDLRDDLRDDL")

func attestTAPublicKey(publicKey *secp256k1.PublicKey) {

	var pub_hex_string string = hex.EncodeToString(publicKey.SerializeCompressed())
	var ta string = "'{\"pubkey\": \"" + pub_hex_string + "\"}'"
	cmd := exec.Command("bash", "-c", "/home/jeckel/go/bin/planetmint-god tx machine register-trust-anchor "+ta+" --from cosmos10hyme8ggv30q6mru7zfeleryac4vm3xs26n0ft -y")
	out, err := cmd.Output()
	if err != nil {
		// if there was any error, print it here
		fmt.Println("could not run command: ", err)
	}
	// otherwise, print the output from running the command
	fmt.Println("Output: ", string(out))
}

// getAlbums responds with the list of all albums as JSON.
func getFirmware(c *gin.Context) {
	privKey, pubKey := generateNewKeyPair()

	var patched_binary = bytes.Replace(reference_content, searchBytes, privKey.Serialize(), 1)
	c.Header("Content-Disposition", "attachment; filename=firmware.patched")
	c.Data(http.StatusOK, "application/octet-stream", patched_binary)
	fmt.Println(" pub key 1: ", pubKey.SerializeCompressed())
	attestTAPublicKey(pubKey)
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

func main() {
	// pk_source, _ := getRandomPrivateKey(32)

	// privateKeyBytes, err := hex.DecodeString(pk_source)
	// if err != nil {
	// 	log.Fatalf("Failed to decode private key: %v", err)
	// }

	// // Initialize a secp256k1 private key object.
	// privateKey, publicKey := btcec.PrivKeyFromBytes(privateKeyBytes)

	// // Print the public key for verification.
	// fmt.Printf("Public Key: %x\n", publicKey.SerializeCompressed())

	// // Replace this with the digest you want to sign.
	// digest := []byte("Your digest to be signed")

	// // Sign the digest using the private key.
	// //r, signature, err := ecdsa.Sign(privateKey, digest)
	// signature := ecdsa.Sign(privateKey, digest)
	// if err != nil {
	// 	log.Fatalf("Failed to sign digest: %v", err)
	// }

	// // Serialize the signature in a format you prefer.
	// signatureBytes := signature.Serialize()

	// fmt.Printf("Signature: %x\n", signatureBytes)
	// ser := privateKey.Serialize()
	// //obj, err := hex.DecodeString(ser)
	// fmt.Printf("Private Key: %x\n", privateKey.Serialize())
	// fmt.Printf("Private Key: %x\n", ser)

	// fmt.Printf("Public Key (compressed): %x\n", publicKey.SerializeCompressed())
	// fmt.Printf("Public Key (uncompressed): %x\n", publicKey.SerializeUncompressed())

	content, err := os.ReadFile("./tasmota32-rddl.bin")
	if err != nil {
		panic("could not read firmware")
	}
	reference_content = content
	// var ta string = "'{\"pubkey\": \"0254ee8462dd6a8c4ef7bdc21ac92bd222aa53e4284f9d743f8bd196b899f0ada5\"}'"
	// cmd := exec.Command("bash", "-c", "/home/jeckel/go/bin/planetmint-god tx machine register-trust-anchor "+ta+" --from cosmos10hyme8ggv30q6mru7zfeleryac4vm3xs26n0ft -y")
	// out, err := cmd.Output()
	// if err != nil {
	// //if there was any error, print it here
	// fmt.Println("could not run command: ", err)
	// }
	// //otherwise, print the output from running the command
	// fmt.Println("Output: ", string(out))

	router := gin.Default()
	router.GET("/firmware", getFirmware)

	router.Run("localhost:8080")
}
