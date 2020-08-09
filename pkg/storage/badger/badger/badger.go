// Package badgeer provides handy badger interfaces
package badgeer

import (
	"encoding/json"

	"github.com/dgraph-io/badger/v2"
)

func Open(opt badger.Options) (*DB, error) {
	db, err := badger.Open(opt)
	if err != nil {
		return nil, err
	}

	return New(db), nil
}

func New(db *badger.DB) *DB {
	return &DB{db}
}

type DB struct {
	*badger.DB
}

func (db *DB) GetJSON(key string, data interface{}) error {
	if err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		if err := item.Value(func(val []byte) error {
			return json.Unmarshal(val, data)
		}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (db *DB) SetJSON(key string, data interface{}) error {
	if err := db.Update(func(txn *badger.Txn) error {
		value, err := json.Marshal(data)
		if err != nil {
			return err
		}

		if err := txn.Set([]byte(key), value); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (db *DB) Delete(key string) error {
	if err := db.DB.Update(func(txn *badger.Txn) error {
		if err := txn.Delete([]byte(key)); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}
