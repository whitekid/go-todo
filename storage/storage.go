package storage

//go:generate mockgen -destination=mocks/mocks.go . Interface

import (
	"github.com/pkg/errors"
	"github.com/whitekid/go-todo/config"
	"github.com/whitekid/go-todo/storage/badger"
	"github.com/whitekid/go-todo/storage/types"
)

var (
	ErrNotFound         = types.ErrNotFound
	ErrNotAuthenticated = types.ErrNotAuthenticated

	Today = types.Today
)

type (
	Interface   = types.Interface
	TodoStorage = types.TodoService
	User        = types.User

	TodoItem = types.TodoItem
)

// storage factories
var storages = map[string]factoryFunc{
	badger.Name: func(name string) Interface {
		intf, err := badger.New(name)
		if err != nil {
			panic(err)
		}
		return intf
	},
}

type factoryFunc func(name string) Interface

func New(name string) (Interface, error) {
	factory, ok := storages[config.Storage()]
	if !ok {
		return nil, errors.Errorf(`unknown storage type: "%s"`, config.Storage())
	}

	return factory(name), nil
}
