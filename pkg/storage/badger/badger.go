package badger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v2"
	"github.com/pkg/errors"
	badgerx "github.com/whitekid/go-todo/pkg/storage/badger/badger"
	. "github.com/whitekid/go-todo/pkg/storage/types"
	"github.com/whitekid/go-todo/pkg/utils"
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
	db, err := badgerx.Open(badger.DefaultOptions(name + ".db").WithLogger(l))
	if err != nil {
		return nil, errors.Wrap(err, "badger.New")
	}

	ctx, cancel := context.WithCancel(context.TODO())
	s := &badgerStorage{
		cancel: cancel,
		db:     db,
	}

	s.userService = &badgerUserService{
		storage: s,
	}

	s.tokenService = &badgerTokenService{
		storage: s,
	}

	s.todoService = &badgerTodoService{
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
// /users/{email} --> User object
//
type badgerUserService struct {
	storage *badgerStorage
}

func (s *badgerUserService) Create(user *User) error {
	if err := s.storage.db.SetJSON(fmt.Sprintf("/users/%s", user.Email), user); err != nil {
		return errors.Wrapf(err, "user.Create()")
	}

	return nil
}

func (s *badgerUserService) Get(email string) (*User, error) {
	var user User

	if err := s.storage.db.GetJSON(fmt.Sprintf("/users/%s", email), &user); err != nil {
		return nil, err
	}

	return &user, nil
}

//
//  /todos/all/c9be58c3-e164-42de-a301-8ad3fdbf553b    : all todo object
//  /todos/{email}/c9be58c3-e164-42de-a301-8ad3fdbf553b  : users todo item
//
type badgerStorage struct {
	cancel context.CancelFunc
	db     *badgerx.DB

	userService  *badgerUserService
	tokenService *badgerTokenService
	todoService  *badgerTodoService
}

func (s *badgerStorage) Close() {
	s.cancel()
}

func (s *badgerStorage) UserService() UserService {
	return s.userService
}

func (s *badgerStorage) TokenService() TokenService {
	return s.tokenService
}

func (s *badgerStorage) TodoService() TodoService {
	return s.todoService
}

//
// /access_tokens/{access_token} --> AccessToken object
//
type badgerTokenService struct {
	storage *badgerStorage
}

func (s *badgerTokenService) Create(email string) (*AccessToken, error) {
	// check if user exists
	user, err := s.storage.userService.Get(email)
	if err != nil {
		if err == badger.ErrKeyNotFound {
			user = &User{
				Email: email,
			}

			if err := s.storage.userService.Create(user); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	// create token
	token := &AccessToken{
		Email:  user.Email,
		Token:  utils.RandomString(40),
		Expire: time.Now().UTC().AddDate(0, 0, 1),
	}
	if err := s.storage.db.SetJSON(fmt.Sprintf("/access_tokens/%s", token.Token), token); err != nil {
		return nil, err
	}

	return token, nil
}

func (s *badgerTokenService) Get(token string) (*AccessToken, error) {
	var accessToken AccessToken
	if err := s.storage.db.GetJSON(fmt.Sprintf("/access_tokens/%s", token), &accessToken); err != nil {
		return nil, err
	}
	return &accessToken, nil
}

func (s *badgerTokenService) Delete(token string) error {
	return s.storage.db.Delete(fmt.Sprintf("/access_tokens/%s", token))
}

type badgerTodoService struct {
	storage *badgerStorage

	deletedCh chan *string
	updateCh  chan *string
}

func (t *badgerTodoService) keyTodoItem(email, id string) string {
	if email == "" {
		return "/todos/" + email + "/" + id
	} else {
		return "/todos/all/" + id
	}
}

func (t *badgerTodoService) keyAllTodItem(id string) string {
	return "/todos/all/" + id
}

func (t *badgerTodoService) List(email string) ([]TodoItem, error) {
	items := []TodoItem{}

	if err := t.storage.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		prefix := []byte(t.keyTodoItem(email, ""))
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

func (t *badgerTodoService) Create(email string, item *TodoItem) error {
	if err := t.storage.db.SetJSON(t.keyTodoItem(email, item.ID), item); err != nil {
		return err
	}

	return nil
}

func (t *badgerTodoService) Get(email, itemID string) (*TodoItem, error) {
	var todo TodoItem

	if err := t.storage.db.GetJSON(t.keyTodoItem(email, itemID), &todo); err != nil {
		return nil, err
	}
	return &todo, nil
}

func (t *badgerTodoService) Update(email string, item *TodoItem) error {
	if err := t.storage.db.SetJSON(t.keyTodoItem(email, item.ID), item); err != nil {
		return err
	}
	return nil
}

func (t *badgerTodoService) Delete(email string, itemID string) error {
	if err := t.storage.db.Delete(t.keyTodoItem(email, itemID)); err != nil {
		if err == badger.ErrKeyNotFound {
			return ErrNotFound
		}

		return err
	}

	t.deletedCh <- &itemID

	return nil
}

// sync to all items when user item updated
func (t *badgerTodoService) syncAllItems() {
	go func() {
		for itemID := range t.deletedCh {
			t.storage.db.Delete(t.keyAllTodItem(*itemID))
		}
	}()

	go func() {
		for itemID := range t.updateCh {
			var item TodoItem

			if err := t.storage.db.GetJSON(t.keyTodoItem("", *itemID), &item); err != nil {
				log.Errorf("sync update failed: %v", err)
				continue
			}

			if err := t.storage.db.SetJSON(t.keyAllTodItem(*itemID), &item); err != nil {
				log.Errorf("sync update failed: %v", err)
			}
		}
	}()
}
