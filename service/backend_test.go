package service_test

import (
	"strconv"
	"testing"

	"github.com/rddl-network/ta_attest/service"
	"github.com/rddl-network/ta_attest/types"
	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
)

func createNAccounts(n int) (accounts []types.PostFundingRequest) {
	accounts = make([]types.PostFundingRequest, n)
	for i := range accounts {
		accounts[i].MachineID = "machineID/" + strconv.Itoa(i)
		accounts[i].PlmntAddress = "plmntAddr/" + strconv.Itoa(i)
		accounts[i].Signature = "signature/" + strconv.Itoa(i)
	}
	return
}

func TestAccount(t *testing.T) {
	db, err := leveldb.Open(storage.NewMemStorage(), nil)
	assert.NoError(t, err)

	accounts := createNAccounts(10)
	for _, account := range accounts {
		err := service.StoreAccount(db, account)
		assert.NoError(t, err)
	}

	for _, account := range accounts {
		found, err := service.HasAccount(db, account.PlmntAddress)
		assert.True(t, found)
		assert.NoError(t, err)
	}
}
