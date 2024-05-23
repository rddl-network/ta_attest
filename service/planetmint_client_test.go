package service_test

import (
	"log"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/rddl-network/ta_attest/config"
	"github.com/rddl-network/ta_attest/service"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestPlanetmintQueryAccount(t *testing.T) {
	cfg := config.DefaultConfig()
	grpcConn, err := grpc.Dial(
		cfg.PlanetmintRPCHost,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.ForceCodec(codec.NewProtoCodec(nil).GRPCCodec())),
	)
	if err != nil {
		log.Fatalf("fatal error opening grpc connection %s", err)
	}
	cfg.PlanetmintActor = "plmnt1rmgh77rsnfz2vk2tn3j0uqczw97znpy0lxp4sm"
	pmc := service.NewPlanetmintClient(cfg.PlanetmintActor, grpcConn)
	res, err := pmc.GetAccount("plmnt199zf0vkmehhr2hhdt3e425r5dx4749dmenm35w")
	assert.NoError(t, err)
	assert.NotNil(t, res)

}
