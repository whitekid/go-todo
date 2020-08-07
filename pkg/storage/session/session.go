//Package session implements http session based storage
package session

import (
	"bytes"
	"encoding/json"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	. "github.com/whitekid/go-todo/pkg/storage/types"
	. "github.com/whitekid/go-todo/pkg/types"
	log "github.com/whitekid/go-utils/logging"
)

const (
	Name     = "session"
	keyItems = "items"
)

// New create new session based storage
// session storage does not provide user
func New(name string) Interface {
	s := &sessionStorage{}

	s.todoService = &todoStorage{storage: s}
	return s
}

type sessionStorage struct {
	context *Context

	todoService *todoStorage
}

func (s *sessionStorage) SetContext(c echo.Context) {
	if c == nil {
		s.context = nil
	} else {
		s.context = c.(*Context)
	}
}

func (s *sessionStorage) Close() {
}

func (s *sessionStorage) TodoService() TodoStorage {
	return s.todoService
}

type todoStorage struct {
	storage *sessionStorage
}

func (t *todoStorage) List() ([]TodoItem, error) {
	value, ok := t.storage.context.Session().Values[keyItems]
	if !ok {
		value = []byte{}
	}

	items := make([]TodoItem, 0)
	buf, ok := value.([]byte)
	b := bytes.NewBuffer(buf)
	if err := json.NewDecoder(b).Decode(&items); err != nil {
		log.Warnf("json decode failed: %s, buf: %s, reset to empty items", err, string(buf))
	}

	return items, nil
}

func (t *todoStorage) saveItems(items []TodoItem) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(items); err != nil {
		return errors.Wrapf(err, "saveItems")
	}

	sess := t.storage.context.Session()
	sess.Values[keyItems] = buf.Bytes()

	log.Infof("save items %+v, data: %s", items, buf.String())
	return sess.Save(t.storage.context.Request(), t.storage.context.Response())
}

func (t *todoStorage) Create(item *TodoItem) error {
	items, _ := t.List()
	items = append(items, *item)

	return t.saveItems(items)
}

func (t *todoStorage) Get(itemID string) (*TodoItem, error) {
	items, _ := t.List()
	for i, item := range items {
		if item.ID == itemID {
			return &items[i], nil
		}
	}

	return nil, ErrNotFound
}

func (t *todoStorage) Update(item *TodoItem) error {
	items, _ := t.List()
	for i, e := range items {
		if e.ID == item.ID {
			items[i] = *item

			return t.saveItems(items)
		}
	}

	return ErrNotFound
}

func (t *todoStorage) Delete(itemID string) error {
	items, _ := t.List()
	for i, e := range items {
		if e.ID == itemID {
			items := append(items[:i], items[i+1:]...)
			return t.saveItems(items)
		}
	}

	return ErrNotFound
}
