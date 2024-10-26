package service

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rddl-network/go-utils/signature"
	"github.com/rddl-network/ta_attest/types"
	"github.com/syndtr/goleveldb/leveldb"
)

func (s *TAAService) GetRouter() *gin.Engine {
	return s.router
}

func (s *TAAService) postPubKey(c *gin.Context) {
	pubKey := c.Param("pubkey")
	_, err := hex.DecodeString(pubKey)
	if err == nil {
		fmt.Println(" pub key: " + pubKey)
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
		s.logger.Error("msg", err.Error())
		return
	}

	s.logger.Info("msg", "create-account request received", "machineID", requestBody.MachineID, "signature", requestBody.Signature, "plmntAddress", requestBody.PlmntAddress)

	// verify machine ID validity
	isValidSecp256r1, errR1 := signature.ValidateSECP256R1Signature(requestBody.MachineID, requestBody.Signature, requestBody.MachineID)
	if errR1 != nil || !isValidSecp256r1 {
		isValidSecp256k1, errK1 := signature.ValidateSignature(requestBody.MachineID, requestBody.Signature, requestBody.MachineID)
		if errK1 != nil || !isValidSecp256k1 {
			errStr := ""
			if errR1 != nil {
				errStr = errR1.Error() + ", "
			}
			s.logger.Error("msg", errStr+errR1.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": errStr + errK1.Error()})
			return
		}
	}

	// check if account already in db
	found, err := HasAccount(s.db, requestBody.PlmntAddress)
	if err != nil && !errors.Is(err, leveldb.ErrNotFound) {
		s.logger.Error("msg", err.Error())
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
		s.logger.Error("msg", "failed to fetch trust anchor status", "machineID", requestBody.MachineID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to fetch trust anchor status"})
		return
	}

	if taStatus.Isactivated {
		c.JSON(http.StatusBadRequest, gin.H{"error": "trust anchor already in use"})
		return
	}

	// verify plmnt address and not already funded
	account, err := s.pmc.GetAccount(requestBody.PlmntAddress)
	if err != nil {
		s.logger.Error("msg", "failed to fetch account", "plmntAddress", requestBody.PlmntAddress)
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
		s.logger.Error("msg", "failed to send funds ", "address", requestBody.PlmntAddress, "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send funds: " + err.Error()})
		return
	}

	s.logger.Info("msg", "funding successful, storing account", "plmntAddress", requestBody.PlmntAddress, "machineID", requestBody.MachineID)
	err = StoreAccount(s.db, requestBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store account"})
	}
}

func (s *TAAService) GetRoutes() gin.RoutesInfo {
	routes := s.router.Routes()
	return routes
}
