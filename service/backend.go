package service

import (
	"encoding/json"

	"github.com/rddl-network/ta_attest/types"
	"github.com/syndtr/goleveldb/leveldb"
)

func StoreAccount(db *leveldb.DB, req types.PostCreateAccountRequest) (err error) {
	bytes, err := json.Marshal(req)
	if err != nil {
		return
	}
	return db.Put(key(req.PlmntAddress), bytes, nil)
}

func HasAccount(db *leveldb.DB, address string) (found bool, err error) {
	_, err = db.Get(key(address), nil)
	if err != nil {
		return false, err
	}
	return true, nil
}

func key(s string) []byte {
	return []byte(s)
}
