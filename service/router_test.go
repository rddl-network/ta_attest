package service_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/golang/mock/gomock"
	machinetypes "github.com/planetmint/planetmint-go/x/machine/types"
	"github.com/rddl-network/ta_attest/config"
	"github.com/rddl-network/ta_attest/service"
	"github.com/rddl-network/ta_attest/testutil"
	"github.com/rddl-network/ta_attest/types"
	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
)

func TestTestnetModeTrue(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.TestnetMode = true
	db, _ := leveldb.Open(storage.NewMemStorage(), nil)

	ctrl := gomock.NewController(t)
	pmcMock := testutil.NewMockIPlanetmintClient(ctrl)

	s := service.NewTrustAnchorAttestationService(cfg, db, pmcMock)

	routes := s.GetRoutes()
	assert.Equal(t, 3, len(routes))
	assert.Equal(t, "/register/:pubkey", routes[2].Path)
}

func TestTestnetModeFalse(t *testing.T) {
	cfg := config.DefaultConfig()
	db, _ := leveldb.Open(storage.NewMemStorage(), nil)
	ctrl := gomock.NewController(t)
	pmcMock := testutil.NewMockIPlanetmintClient(ctrl)

	s := service.NewTrustAnchorAttestationService(cfg, db, pmcMock)

	routes := s.GetRoutes()
	assert.Equal(t, 2, len(routes))
}

func TestPostCreateAccount(t *testing.T) {
	cfg := config.DefaultConfig()
	db, _ := leveldb.Open(storage.NewMemStorage(), nil)

	defaultMocker := func(t *testing.T) *testutil.MockIPlanetmintClient {
		ctrl := gomock.NewController(t)
		pmcMock := testutil.NewMockIPlanetmintClient(ctrl)
		pmcMock.EXPECT().GetAccount(gomock.Any()).AnyTimes().Return(nil, nil)
		pmcMock.EXPECT().GetTrustAnchorStatus(gomock.Any()).AnyTimes().Return(&machinetypes.QueryGetTrustAnchorStatusResponse{
			Isactivated: false,
		}, nil)
		pmcMock.EXPECT().FundAccount(gomock.Any()).AnyTimes().Return(nil)
		return pmcMock
	}

	tests := []struct {
		desc    string
		reqBody types.PostCreateAccountRequest
		resBody string
		code    int
		mocker  func(t *testing.T) *testutil.MockIPlanetmintClient
	}{
		{
			desc: "valid request",
			reqBody: types.PostCreateAccountRequest{
				MachineID:    "0338F76336CA7F39899476F58B4EC5BFAB93DE4A850479FB6B4FE2EA306072937E",
				Signature:    "DE97AEF2A99B9371882C4639A607A11AF2BA8AE520FF7B28203193F5EB63AE1670D431960C3103682901A8F5B3C542139DCF8FB44F97780FC8D8A45F8A4E59E3",
				PlmntAddress: "plmntAddr",
			},
			resBody: "",
			code:    200,
			mocker:  defaultMocker,
		},
		{
			desc: "invalid signature",
			reqBody: types.PostCreateAccountRequest{
				MachineID:    "02328de87896b9cbb5101c335f40029e4be898988b470abbf683f1a0b318d73470",
				Signature:    "b7479edbf523c55f771991393fce6b481edfab4c85adf60eb12beb5fdc9c13aaa67852d8cd427a3dd5b90f0e4f9bda694e453ef624ff0cde254c3b7ccc7cdcd3",
				PlmntAddress: "plmntAddr",
			},
			resBody: "{\"error\":\"invalid signature\"}",
			code:    400,
			mocker:  defaultMocker,
		},
		{
			desc: "account already funded",
			reqBody: types.PostCreateAccountRequest{
				MachineID:    "0338F76336CA7F39899476F58B4EC5BFAB93DE4A850479FB6B4FE2EA306072937E",
				Signature:    "DE97AEF2A99B9371882C4639A607A11AF2BA8AE520FF7B28203193F5EB63AE1670D431960C3103682901A8F5B3C542139DCF8FB44F97780FC8D8A45F8A4E59E3",
				PlmntAddress: "plmntAddr",
			},
			resBody: "{\"error\":\"account has already been funded\"}",
			code:    400,
			mocker:  defaultMocker,
		},
		{
			desc: "trust anchor already in use",
			reqBody: types.PostCreateAccountRequest{
				MachineID:    "0338F76336CA7F39899476F58B4EC5BFAB93DE4A850479FB6B4FE2EA306072937E",
				Signature:    "DE97AEF2A99B9371882C4639A607A11AF2BA8AE520FF7B28203193F5EB63AE1670D431960C3103682901A8F5B3C542139DCF8FB44F97780FC8D8A45F8A4E59E3",
				PlmntAddress: "otherPlmntAddr",
			},
			resBody: "{\"error\":\"trust anchor already in use\"}",
			code:    400,
			mocker: func(t *testing.T) *testutil.MockIPlanetmintClient {
				ctrl := gomock.NewController(t)
				pmcMock := testutil.NewMockIPlanetmintClient(ctrl)
				pmcMock.EXPECT().GetTrustAnchorStatus(gomock.Any()).Times(1).Return(&machinetypes.QueryGetTrustAnchorStatusResponse{
					Isactivated: true,
				}, nil)
				return pmcMock
			},
		},
		{
			desc: "planetmint account already exists",
			reqBody: types.PostCreateAccountRequest{
				MachineID:    "0338F76336CA7F39899476F58B4EC5BFAB93DE4A850479FB6B4FE2EA306072937E",
				Signature:    "DE97AEF2A99B9371882C4639A607A11AF2BA8AE520FF7B28203193F5EB63AE1670D431960C3103682901A8F5B3C542139DCF8FB44F97780FC8D8A45F8A4E59E3",
				PlmntAddress: "otherPlmntAddr",
			},
			resBody: "{\"error\":\"planetmint account already exists\"}",
			code:    400,
			mocker: func(t *testing.T) *testutil.MockIPlanetmintClient {
				ctrl := gomock.NewController(t)
				pmcMock := testutil.NewMockIPlanetmintClient(ctrl)
				pmcMock.EXPECT().GetAccount(gomock.Any()).Times(1).Return(&authtypes.QueryAccountResponse{}, nil)
				pmcMock.EXPECT().GetTrustAnchorStatus(gomock.Any()).Times(1).Return(&machinetypes.QueryGetTrustAnchorStatusResponse{
					Isactivated: false,
				}, nil)
				return pmcMock
			},
		},
		{
			desc: "failed to send funds",
			reqBody: types.PostCreateAccountRequest{
				MachineID:    "0338F76336CA7F39899476F58B4EC5BFAB93DE4A850479FB6B4FE2EA306072937E",
				Signature:    "DE97AEF2A99B9371882C4639A607A11AF2BA8AE520FF7B28203193F5EB63AE1670D431960C3103682901A8F5B3C542139DCF8FB44F97780FC8D8A45F8A4E59E3",
				PlmntAddress: "otherPlmntAddr",
			},
			resBody: "{\"error\":\"failed to send funds\"}",
			code:    500,
			mocker: func(t *testing.T) *testutil.MockIPlanetmintClient {
				ctrl := gomock.NewController(t)
				pmcMock := testutil.NewMockIPlanetmintClient(ctrl)
				pmcMock.EXPECT().GetAccount(gomock.Any()).Times(1).Return(nil, nil)
				pmcMock.EXPECT().GetTrustAnchorStatus(gomock.Any()).Times(1).Return(&machinetypes.QueryGetTrustAnchorStatusResponse{
					Isactivated: false,
				}, nil)
				pmcMock.EXPECT().FundAccount(gomock.Any()).Times(1).Return(errors.New("some err"))
				return pmcMock
			},
		},
	}

	for i := 0; i < len(tests); i++ {
		tc := tests[i]
		t.Run(tc.desc, func(t *testing.T) {
			pmcMock := tc.mocker(t)
			s := service.NewTrustAnchorAttestationService(cfg, db, pmcMock)
			router := s.GetRouter()

			w := httptest.NewRecorder()
			bodyBytes, err := json.Marshal(tc.reqBody)
			assert.NoError(t, err)
			req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "/create-account", bytes.NewBuffer(bodyBytes))
			assert.NoError(t, err)
			router.ServeHTTP(w, req)
			assert.Equal(t, tc.code, w.Code)
			assert.Equal(t, tc.resBody, w.Body.String())
		})
	}
}
