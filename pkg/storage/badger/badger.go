package badger

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dgraph-io/badger/v2"
	"github.com/pkg/errors"
	badgerx "github.com/whitekid/go-todo/pkg/storage/badger/badger"
	. "github.com/whitekid/go-todo/pkg/storage/types"
	"github.com/whitekid/go-todo/pkg/tokens"
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

		userDeleteCh:  make(chan *string),
		todoDeletedCh: make(chan *string),
		todoUpdateCh:  make(chan *todoUpdate),
	}

	s.userService = &badgerUserService{
		storage: s,
	}

	s.tokenService = &badgerTokenService{
		storage: s,
	}

	s.todoService = &badgerTodoService{
		storage: s,
	}
	go s.handleUpdates()

	// close() callback
	go func() {
		<-ctx.Done()

		close(s.userDeleteCh)
		close(s.todoDeletedCh)
		close(s.todoUpdateCh)

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

func (s *badgerUserService) Delete(email string) error {
	if err := s.storage.db.Delete(fmt.Sprintf("/users/%s", email)); err != nil {
		return errors.Wrapf(err, "delete")
	}

	s.storage.userDeleteCh <- &email

	return nil
}

type todoUpdate struct {
	email *string
	id    *string
}

//
//  /todos/all/c9be58c3-e164-42de-a301-8ad3fdbf553b    : all todo object
//  /todos/{email}/c9be58c3-e164-42de-a301-8ad3fdbf553b  : users todo item
//
type badgerStorage struct {
	cancel context.CancelFunc
	db     *badgerx.DB

	userDeleteCh  chan *string
	todoDeletedCh chan *string
	todoUpdateCh  chan *todoUpdate

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

func (s *badgerStorage) handleUpdates() {
	go func() {
		for email := range s.userDeleteCh {
			// delete user token
			tokensToDelete := []string{}

			s.db.Iter("/tokens/", func(key string, value []byte) error {
				issuer, err := tokens.Parse(string(value))
				if err != nil {
					tokensToDelete = append(tokensToDelete, string(value))
					return nil
				}

				if *email == issuer {
					tokensToDelete = append(tokensToDelete, string(value))
					return nil
				}

				return nil
			})

			for _, t := range tokensToDelete {
				s.tokenService.Delete(t)
			}

			// delete user todo
			ids := []string{}
			s.db.Iter(fmt.Sprintf("/todos/%s/", *email), func(key string, value []byte) error {
				var todo TodoItem
				if err := s.db.GetJSON(key, &todo); err != nil {
					return err
				}
				ids = append(ids, todo.ID)
				return nil
			})
			for _, id := range ids {
				s.todoService.Delete(*email, id)
				s.todoService.Delete("", id)
			}
		}
	}()

	go func() {
		for itemID := range s.todoDeletedCh {
			s.db.Delete(s.todoService.keyTodoItem("", *itemID))
		}
	}()

	go func() {
		for updates := range s.todoUpdateCh {
			var item TodoItem

			// sync user todo --> all todo
			if err := s.db.GetJSON(s.todoService.keyTodoItem(*updates.email, *updates.id), &item); err != nil {
				log.Errorf("sync update failed: %v", err)
				continue
			}

			if err := s.db.SetJSON(s.todoService.keyTodoItem("", *updates.id), &item); err != nil {
				log.Errorf("sync update failed: %v", err)
			}
		}
	}()
}

//
// /tokens/{refresh_tokens} --> RefreshToken object
//
type badgerTokenService struct {
	storage *badgerStorage
}

func (s *badgerTokenService) Create(email, refreshToken string) error {
	// check if user exists
	user, err := s.storage.userService.Get(email)
	if err != nil {
		if err == badger.ErrKeyNotFound {
			user = &User{
				Email: email,
			}

			if err := s.storage.userService.Create(user); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// create token
	if err := s.storage.db.SetString(fmt.Sprintf("/tokens/%s", refreshToken), refreshToken); err != nil {
		return err
	}

	return nil
}

func (s *badgerTokenService) Get(token string) (string, error) {
	t, err := s.storage.db.GetString(fmt.Sprintf("/tokens/%s", token))
	if err != nil {
		return "", err
	}

	if t != token {
		log.Error("token key and data mismatch: key=%s, token=%s", token, t)
		return "", errors.New("token missmatch")
	}

	return t, nil
}

func (s *badgerTokenService) Delete(token string) error {
	if err := s.storage.db.Delete(fmt.Sprintf("/tokens/%s", token)); err != nil {
		return err
	}

	return nil
}

type badgerTodoService struct {
	storage *badgerStorage
}

func (t *badgerTodoService) keyTodoItem(email, id string) string {
	if email == "" {
		return "/todos/" + email + "/" + id
	} else {
		return "/todos/all/" + id
	}
}

func (t *badgerTodoService) List(email string) ([]TodoItem, error) {
	items := []TodoItem{}

	if err := t.storage.db.Iter(t.keyTodoItem(email, ""), func(key string, value []byte) error {
		var item TodoItem

		if err := json.Unmarshal(value, &item); err != nil {
			return err
		}

		if !strings.HasSuffix(string(key), "/"+item.ID) {
			return errors.Errorf("key and value mismatch: key=%s, value=%v", key, item)
		}

		items = append(items, item)
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

	t.storage.todoUpdateCh <- &todoUpdate{&email, &item.ID}

	return nil
}

func (t *badgerTodoService) Delete(email string, itemID string) error {
	if err := t.storage.db.Delete(t.keyTodoItem(email, itemID)); err != nil {
		if err == badger.ErrKeyNotFound {
			return ErrNotFound
		}

		return err
	}

	t.storage.todoDeletedCh <- &itemID

	return nil
}
