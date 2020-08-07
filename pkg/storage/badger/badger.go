package badger

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"

	"github.com/dgraph-io/badger/v2"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	. "github.com/whitekid/go-todo/pkg/storage/types"
	. "github.com/whitekid/go-todo/pkg/types"
	log "github.com/whitekid/go-utils/logging"
)

const Name = "badger"

type logger struct {
	log.Interface
}

func (l *logger) Warningf(fmt string, args ...interface{}) { l.Warnf(fmt, args...) }

// New create new badger storage
func New(name string) (Interface, error) {
	l := &logger{
		Interface: log.New(),
	}
	db, err := badger.Open(badger.DefaultOptions(name + ".db").WithLogger(l))
	if err != nil {
		return nil, errors.Wrap(err, "badger.New")
	}

	ctx, cancel := context.WithCancel(context.TODO())
	s := &badgerStorage{
		cancel: cancel,
		db:     db,
	}
	s.todoService = &badgerTodoStorage{
		storage:   s,
		deletedCh: make(chan *string),
		updateCh:  make(chan *string),
	}
	go s.todoService.syncAllItems()

	// close() callback
	go func() {
		<-ctx.Done()
		close(s.todoService.deletedCh)
		close(s.todoService.updateCh)

		s.db.Close()
	}()
	return s, nil
}

//
//  /todos/all/c9be58c3-e164-42de-a301-8ad3fdbf553b    : all todo object
//  /todos/{email}/c9be58c3-e164-42de-a301-8ad3fdbf553b  : users todo item
//
type badgerStorage struct {
	cancel context.CancelFunc
	db     *badger.DB
	email  string

	todoService *badgerTodoStorage
}

func (s *badgerStorage) SetContext(c echo.Context) {
	if c == nil {
		s.email = ""
	} else {
		s.email = c.(*Context).Email()
	}
}

func (s *badgerStorage) Close() {
	s.cancel()
}

func (s *badgerStorage) TodoService() TodoStorage {
	return s.todoService
}

type badgerTodoStorage struct {
	storage *badgerStorage

	deletedCh chan *string
	updateCh  chan *string
}

func (t *badgerTodoStorage) keyTodoItem(id string) []byte {
	return []byte("/todos/" + t.storage.email + "/" + id)
}

func (t *badgerTodoStorage) keyAllTodItem(id string) []byte {
	return []byte("/todos/all/" + id)
}

func (t *badgerTodoStorage) List() ([]TodoItem, error) {
	if t.storage.email == "" {
		return nil, ErrNotAuthenticated
	}

	items := []TodoItem{}

	if err := t.storage.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		prefix := []byte(t.keyTodoItem(""))
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			key := item.Key()
			if err := item.Value(func(v []byte) error {
				var item TodoItem

				buf := bytes.NewBuffer(v)
				if err := json.NewDecoder(buf).Decode(&item); err != nil {
					return err
				}

				if !strings.HasSuffix(string(key), "/"+item.ID) {
					return errors.Errorf("key and value mismatch: key=%s, value=%v", key, item)
				}

				items = append(items, item)

				return nil
			}); err != nil {
				return err
			}
			return nil
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return items, nil
}

func (t *badgerTodoStorage) Create(item *TodoItem) error {
	if t.storage.email == "" {
		return ErrNotAuthenticated
	}

	if err := t.storage.db.Update(func(txn *badger.Txn) error {
		value, err := json.Marshal(item)
		if err != nil {
			return err
		}

		if err := txn.Set(t.keyTodoItem(item.ID), value); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (t *badgerTodoStorage) Get(itemID string) (*TodoItem, error) {
	if t.storage.email == "" {
		return nil, ErrNotAuthenticated
	}

	var todo TodoItem

	if err := t.storage.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(t.keyTodoItem(itemID))
		if err != nil {
			return err
		}

		if err := item.Value(func(v []byte) error {
			if err := json.Unmarshal(v, &todo); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}

		return nil
	}); err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &todo, nil
}

func (t *badgerTodoStorage) Update(item *TodoItem) error {
	if t.storage.email == "" {
		return ErrNotAuthenticated
	}

	if err := t.storage.db.Update(func(txn *badger.Txn) error {
		value, err := json.Marshal(item)
		if err != nil {
			return err
		}

		if err := txn.Set(t.keyTodoItem(item.ID), value); err != nil {
			return err
		}

		return nil
	}); err != nil {
		if err == badger.ErrKeyNotFound {
			return ErrNotFound
		}

		return err
	}

	return nil
}

func (t *badgerTodoStorage) Delete(itemID string) error {
	if t.storage.email == "" {
		return ErrNotAuthenticated
	}

	if err := t.storage.db.Update(func(txn *badger.Txn) error {
		if err := txn.Delete(t.keyTodoItem(itemID)); err != nil {
			return err
		}

		return nil
	}); err != nil {
		if err == badger.ErrKeyNotFound {
			return ErrNotFound
		}

		return err
	}

	t.deletedCh <- &itemID

	return nil
}

// sync to all items when user item updated
func (t *badgerTodoStorage) syncAllItems() {
	go func() {
		for itemID := range t.deletedCh {
			t.storage.db.Update(func(txn *badger.Txn) error {
				if err := txn.Delete(t.keyAllTodItem(*itemID)); err != nil {
					return err
				}

				return nil
			})
		}
	}()

	go func() {
		for itemID := range t.updateCh {
			var value []byte

			if err := t.storage.db.View(func(txn *badger.Txn) error {
				item, err := txn.Get(t.keyTodoItem(*itemID))
				if err != nil {
					return err
				}

				if err := item.Value(func(v []byte) error {
					value = v
					return nil
				}); err != nil {
					return err
				}

				return nil
			}); err != nil {
				log.Errorf("sync update failed: %v", err)
				continue
			}

			t.storage.db.Update(func(txn *badger.Txn) error {
				if err := txn.Set(t.keyTodoItem(*itemID), value); err != nil {
					return err
				}

				return nil
			})
		}
	}()
}
