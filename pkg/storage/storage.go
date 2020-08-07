package storage

import (
	"github.com/pkg/errors"
	"github.com/whitekid/go-todo/pkg/config"
	"github.com/whitekid/go-todo/pkg/storage/badger"
	"github.com/whitekid/go-todo/pkg/storage/session"
	"github.com/whitekid/go-todo/pkg/storage/types"
)

var (
	ErrNotFound = types.ErrNotFound
	Today       = types.Today
)

type (
	Interface   = types.Interface
	TodoStorage = types.TodoStorage

	TodoItem = types.TodoItem
)

var storages = map[string]Factory{
	session.Name: func(name string) Interface {
		return session.New(name)
	},
	badger.Name: func(name string) Interface {
		intf, err := badger.New(name)
		if err != nil {
			panic(err)
		}
		return intf
	},
}

type Factory func(name string) Interface

func New(name string) (Interface, error) {
	factory, ok := storages[config.Storage()]
	if !ok {
		return nil, errors.Errorf("unknown storage: " + config.Storage())
	}

	return factory(name), nil
}
