package service_test

import (
	"io"
	"net/http"
	"testing"

	"github.com/rddl-network/ta_attest/service"
	"github.com/stretchr/testify/assert"
)

func TestFirmwareHandling(t *testing.T) {
	privKey, _ := service.GenerateNewKeyPair()
	firmwareURL := "https://github.com/rddl-network/Tasmota/releases/download/rddl-v0.40.1/tasmota32-rddl.bin"

	resp, err := http.Get(firmwareURL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	defer resp.Body.Close()

	firmware, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	valid := service.VerifyBinaryIntegrity(firmware)
	assert.True(t, valid)

	patchedFirmware := service.PatchFirmware(firmware, privKey)
	invalid := service.VerifyBinaryIntegrity(patchedFirmware)
	assert.False(t, invalid)

	service.ComputeAndSetFirmwareChecksum(patchedFirmware)
	valid = service.VerifyBinaryIntegrity(patchedFirmware)
	assert.True(t, valid)
}
