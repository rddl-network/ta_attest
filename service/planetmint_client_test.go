package service_test

import (
	"log"
	"testing"

	"github.com/rddl-network/ta_attest/config"
	"github.com/rddl-network/ta_attest/service"
	"github.com/stretchr/testify/assert"
)

func TestPlanetmintQueryAccount(t *testing.T) {
	// skipped because test is just to showcase interfaceRegistry fix
	t.SkipNow()
	cfg := config.DefaultConfig()
	grpcConn, err := service.SetupGRPCConnection(cfg)
	if err != nil {
		log.Fatalf("fatal error opening grpc connection %s", err)
	}
	cfg.PlanetmintActor = "plmnt1p445cz0hfg4yg3dgrq5n3e9wdr8rwpt9qfcz2y"
	pmc := service.NewPlanetmintClient(cfg.PlanetmintActor, grpcConn)
	res, err := pmc.GetAccount("plmnt1u8awp62tsp68fed7ezasdh98rch6wyansg8dzs")
	assert.NoError(t, err)
	assert.NotNil(t, res)
}
