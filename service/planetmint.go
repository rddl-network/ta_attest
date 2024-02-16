package service

import (
	"encoding/hex"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/planetmint/planetmint-go/app"
	"github.com/planetmint/planetmint-go/lib"
	machinetypes "github.com/planetmint/planetmint-go/x/machine/types"
)

var libConfig *lib.Config

func init() {
	encodingConfig := app.MakeEncodingConfig()

	libConfig = lib.GetConfig()
	libConfig.SetEncodingConfig(encodingConfig)
}

func (s *TAAService) attestTAPublicKeyHex(pubHexString string) error {
	addr := sdk.MustAccAddressFromBech32(s.cfg.PlanetmintActor)
	msg := machinetypes.NewMsgRegisterTrustAnchor(s.cfg.PlanetmintActor, &machinetypes.TrustAnchor{
		Pubkey: pubHexString,
	})

	_, err := lib.BroadcastTxWithFileLock(addr, msg)
	if err != nil {
		return err
	}
	return nil
}

func (s *TAAService) attestTAPublicKey(publicKey *secp256k1.PublicKey) error {
	pubHexString := hex.EncodeToString(publicKey.SerializeCompressed())
	return s.attestTAPublicKeyHex(pubHexString)
}
