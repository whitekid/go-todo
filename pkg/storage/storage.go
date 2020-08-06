package storage

import (
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

func New() Interface {
	return session.New(nil, nil)
}
