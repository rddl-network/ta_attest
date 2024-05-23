package service

import (
	"context"
	"encoding/hex"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	ctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/planetmint/planetmint-go/app"
	"github.com/planetmint/planetmint-go/lib"
	machinetypes "github.com/planetmint/planetmint-go/x/machine/types"
	"github.com/rddl-network/ta_attest/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
)

type IPlanetmintClient interface {
	AttestTAPublicKey(publicKey *secp256k1.PublicKey) error
	AttestTAPublicKeyHex(pubHexString string) error
	GetTrustAnchorStatus(machineID string) (res *machinetypes.QueryGetTrustAnchorStatusResponse, err error)
	GetAccount(plmntAddress string) (res *authtypes.QueryAccountResponse, err error)
	FundAccount(plmntAddress string) error
}

type PlanetmintClient struct {
	actor string
	conn  *grpc.ClientConn
}

func NewPlanetmintClient(actor string, conn *grpc.ClientConn) *PlanetmintClient {
	return &PlanetmintClient{
		actor: actor,
		conn:  conn,
	}
}

func SetupGRPCConnection(cfg *config.Config) (conn *grpc.ClientConn, err error) {
	interfaceRegistry := ctypes.NewInterfaceRegistry()
	interfaceRegistry.RegisterInterface(
		"cosmos.auth.IAccount",
		(*authtypes.AccountI)(nil),
		&authtypes.BaseAccount{},
		&authtypes.ModuleAccount{},
	)

	return grpc.Dial(
		cfg.PlanetmintRPCHost,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.ForceCodec(codec.NewProtoCodec(interfaceRegistry).GRPCCodec())),
	)
}

var libConfig *lib.Config

func init() {
	encodingConfig := app.MakeEncodingConfig()

	libConfig = lib.GetConfig()
	libConfig.SetEncodingConfig(encodingConfig)
}

func (pmc *PlanetmintClient) AttestTAPublicKeyHex(pubHexString string) error {
	addr := sdk.MustAccAddressFromBech32(pmc.actor)
	msg := machinetypes.NewMsgRegisterTrustAnchor(pmc.actor, &machinetypes.TrustAnchor{
		Pubkey: pubHexString,
	})

	_, err := lib.BroadcastTxWithFileLock(addr, msg)
	if err != nil {
		return err
	}
	return nil
}

func (pmc *PlanetmintClient) AttestTAPublicKey(publicKey *secp256k1.PublicKey) error {
	pubHexString := hex.EncodeToString(publicKey.SerializeCompressed())
	return pmc.AttestTAPublicKeyHex(pubHexString)
}

func (pmc *PlanetmintClient) GetTrustAnchorStatus(machineID string) (res *machinetypes.QueryGetTrustAnchorStatusResponse, err error) {
	machineClient := machinetypes.NewQueryClient(pmc.conn)
	return machineClient.GetTrustAnchorStatus(
		context.Background(),
		&machinetypes.QueryGetTrustAnchorStatusRequest{Machineid: machineID},
	)
}

func (pmc *PlanetmintClient) GetAccount(plmntAddress string) (res *authtypes.QueryAccountResponse, err error) {
	authClient := authtypes.NewQueryClient(pmc.conn)
	res, err = authClient.Account(
		context.Background(),
		&authtypes.QueryAccountRequest{Address: plmntAddress},
	)

	if err != nil {
		if strings.Contains(err.Error(), codes.NotFound.String()) {
			err = nil
		}
		return
	}
	return
}

func (pmc *PlanetmintClient) FundAccount(plmntAddress string) error {
	fromAddr := sdk.MustAccAddressFromBech32(pmc.actor)
	toAddr := sdk.MustAccAddressFromBech32(plmntAddress)
	msg := banktypes.NewMsgSend(
		fromAddr,
		toAddr,
		sdk.NewCoins(sdk.NewCoin("plmnt", sdk.NewInt(1))),
	)

	_, err := lib.BroadcastTxWithFileLock(fromAddr, msg)
	return err
}
