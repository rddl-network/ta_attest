package service

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	pkiutil "github.com/rddl-network/go-utils/pki"
	"github.com/rddl-network/go-utils/signature"
	"github.com/rddl-network/ta_attest/types"
	"github.com/syndtr/goleveldb/leveldb"
)

func (s *TAAService) GetRouter() *gin.Engine {
	return s.router
}

func (s *TAAService) getFirmware(c *gin.Context) {
	mcu := c.Param("mcu")
	pkSource, err := pkiutil.GetRandomPrivateKey()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("%w", err)})
		return
	}
	privKey, pubKey, err := pkiutil.GenerateNewKeyPair(pkSource)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("%w", err)})
		return
	}
	var filename string
	var firmwareBytes []byte
	switch mcu {
	case "esp32":
		firmwareBytes = s.firmwareESP32
		filename = "tasmota32-rddl.bin"
	case "esp32c3":
		firmwareBytes = s.firmwareESP32C3
		filename = "tasmota32c3-rddl.bin"
	default:
		c.String(404, "Resource not found, Firmware not supported")
		return
	}

	patchedFirmware := PatchFirmware(firmwareBytes, privKey)
	ComputeAndSetFirmwareChecksum(patchedFirmware)

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, "application/octet-stream", patchedFirmware)

	fmt.Println(" pub key: ", hex.EncodeToString(pubKey.SerializeCompressed()))
	_ = s.pmc.AttestTAPublicKey(pubKey)
}

func (s *TAAService) postPubKey(c *gin.Context) {
	pubKey := c.Param("pubkey")
	_, err := hex.DecodeString(pubKey)
	if err == nil {
		err = s.pmc.AttestTAPublicKeyHex(pubKey)
		if err == nil {
			c.IndentedJSON(http.StatusOK, pubKey)
		} else {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		}
	} else {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid pubkey"})
	}
}

func (s *TAAService) createAccount(c *gin.Context) {
	var requestBody types.PostCreateAccountRequest
	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// verify machine ID validity
	isValid, err := signature.ValidateSignature(requestBody.MachineID, requestBody.Signature, requestBody.MachineID)
	if err != nil || !isValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// check if account already in db
	found, err := HasAccount(s.db, requestBody.PlmntAddress)
	if err != nil && !errors.Is(err, leveldb.ErrNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read db"})
		return
	}

	if found {
		c.JSON(http.StatusBadRequest, gin.H{"error": "account has already been funded"})
		return
	}

	// verify trust anchor registered
	taStatus, err := s.pmc.GetTrustAnchorStatus(requestBody.MachineID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch trust anchor status"})
		return
	}

	if taStatus.Isactivated {
		c.JSON(http.StatusBadRequest, gin.H{"error": "trust anchor already in use"})
		return
	}

	// verify plmnt address and not already funded
	account, err := s.pmc.GetAccount(requestBody.PlmntAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch account"})
		return
	}

	// If account exists no need for funding
	if account != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "planetmint account already exists"})
		return
	}

	err = s.pmc.FundAccount(requestBody.PlmntAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send funds"})
		return
	}

	err = StoreAccount(s.db, requestBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store account"})
	}
}

func (s *TAAService) GetRoutes() gin.RoutesInfo {
	routes := s.router.Routes()
	return routes
}
