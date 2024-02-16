package service

import (
	"crypto/rand"
	"encoding/hex"
	"log"

	btcec "github.com/btcsuite/btcd/btcec/v2"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
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

func GenerateNewKeyPair() (*secp256k1.PrivateKey, *secp256k1.PublicKey) {
	pkSource, _ := getRandomPrivateKey(32)
	privateKeyBytes, err := hex.DecodeString(pkSource)
	if err != nil {
		log.Fatalf("Failed to decode private key: %v", err)
	}

	// Initialize a secp256k1 private key object.
	privateKey, publicKey := btcec.PrivKeyFromBytes(privateKeyBytes)
	return privateKey, publicKey
}
