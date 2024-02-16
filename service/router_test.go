package service_test

import (
	"testing"

	"github.com/rddl-network/ta_attest/config"
	"github.com/rddl-network/ta_attest/service"
	"gotest.tools/assert"
)

func TestTestnetModeTrue(t *testing.T) {
	t.Parallel()
	cfg := config.DefaultConfig()
	cfg.TestnetMode = true
	s := service.NewTrustAnchorAttestationService(cfg)

	routes := s.GetRoutes()
	assert.Equal(t, 2, len(routes))
	assert.Equal(t, "/register/:pubkey", routes[1].Path)
}

func TestTestnetModeFalse(t *testing.T) {
	t.Parallel()
	cfg := config.DefaultConfig()
	s := service.NewTrustAnchorAttestationService(cfg)

	routes := s.GetRoutes()
	assert.Equal(t, 1, len(routes))
}
