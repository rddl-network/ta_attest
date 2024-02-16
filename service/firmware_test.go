package service_test

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/rddl-network/ta_attest/service"
	"github.com/stretchr/testify/assert"
)

func TestFirmwareHandling(t *testing.T) {
	t.Parallel()
	privKey, _ := service.GenerateNewKeyPair()
	firmwareURL := "https://github.com/rddl-network/Tasmota/releases/download/rddl-v0.40.1/tasmota32-rddl.bin"

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // 30 seconds timeout
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, firmwareURL, nil)
	assert.NoError(t, err)

	client := &http.Client{}
	resp, err := client.Do(req)
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
