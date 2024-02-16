package service

import (
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *TAAService) getFirmware(c *gin.Context) {
	mcu := c.Param("mcu")
	privKey, pubKey := GenerateNewKeyPair()
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
	_ = s.attestTAPublicKey(pubKey)
}

func (s *TAAService) postPubKey(c *gin.Context) {
	pubKey := c.Param("pubkey")
	_, err := hex.DecodeString(pubKey)
	if err == nil {
		err = s.attestTAPublicKeyHex(pubKey)
		if err == nil {
			c.IndentedJSON(http.StatusOK, pubKey)
		} else {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		}
	} else {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid pubkey"})
	}
}

func (s *TAAService) GetRoutes() gin.RoutesInfo {
	routes := s.router.Routes()
	return routes
}
