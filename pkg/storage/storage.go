package storage

import (
	"github.com/pkg/errors"
	"github.com/whitekid/go-todo/pkg/config"
	"github.com/whitekid/go-todo/pkg/storage/badger"
	"github.com/whitekid/go-todo/pkg/storage/types"
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

var storages = map[string]Factory{
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
		return nil, errors.Errorf(`unknown storage: "%s"`, config.Storage())
	}

	return factory(name), nil
}
