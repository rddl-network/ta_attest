package service

import (
	"encoding/json"

	"github.com/rddl-network/ta_attest/types"
	"github.com/syndtr/goleveldb/leveldb"
)

func initDB(path string) (db *leveldb.DB) {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		panic("error while initializing database: " + err.Error())
	}
	return db
}

func StoreAccount(db *leveldb.DB, req types.PostFundingRequest) (err error) {
	bytes, err := json.Marshal(req)
	if err != nil {
		return
	}
	db.Put(key(req.PlmntAddress), bytes, nil)
	return
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
