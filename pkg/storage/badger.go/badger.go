package badger

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/dgraph-io/badger/v2"
	"github.com/pkg/errors"
	. "github.com/whitekid/go-todo/pkg/storage/types"
	"github.com/whitekid/go-utils/logging"
)

type logger struct {
	logging.Interface
}

func (l *logger) Warningf(fmt string, args ...interface{}) { l.Warnf(fmt, args...) }

// New create new badger storage
func New(name string) (Interface, error) {
	l := &logger{
		Interface: logging.New(),
	}
	db, err := badger.Open(badger.DefaultOptions(name).WithLogger(l))
	if err != nil {
		return nil, errors.Wrap(err, "badger.New")
	}

	s := &badgerStorage{db: db}
	s.todoService = &badgerTodoStorage{storage: s}
	return s, nil
}

//
//  /todos/c9be58c3-e164-42de-a301-8ad3fdbf553b : json object
//
type badgerStorage struct {
	db *badger.DB

	todoService *badgerTodoStorage
}

func (s *badgerStorage) TodoService() TodoStorage {
	return s.todoService
}

type badgerTodoStorage struct {
	storage *badgerStorage
}

func keyTodoItem(id string) []byte {
	return []byte("/todos/" + id)
}

func (t *badgerTodoStorage) List() ([]TodoItem, error) {
	items := []TodoItem{}

	if err := t.storage.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		prefix := []byte("/todos/")
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
	if err := t.storage.db.Update(func(txn *badger.Txn) error {
		value, err := json.Marshal(item)
		if err != nil {
			return err
		}

		if err := txn.Set(keyTodoItem(item.ID), value); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (t *badgerTodoStorage) Get(itemID string) (*TodoItem, error) {
	var todo TodoItem

	if err := t.storage.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(keyTodoItem(itemID))
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
		return nil, err
	}

	return &todo, nil
}

func (t *badgerTodoStorage) Update(item *TodoItem) error {
	if err := t.storage.db.Update(func(txn *badger.Txn) error {
		value, err := json.Marshal(item)
		if err != nil {
			return err
		}

		if err := txn.Set(keyTodoItem(item.ID), value); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (t *badgerTodoStorage) Delete(itemID string) error {
	return t.storage.db.Update(func(txn *badger.Txn) error {
		if err := txn.Delete(keyTodoItem(itemID)); err != nil {
			return err
		}

		return nil
	})
}
