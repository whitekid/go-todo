package client

import (
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/whitekid/go-todo/models"
	"github.com/whitekid/go-utils/request"
)

// Interface represents client interface
//go:generate mockgen -destination=mocks/mocks.go -package mocks . Interface
type Interface interface {
	TodoService() TodoService
}

// TodoService ...
type TodoService interface {
	Create(item *models.Item) (*models.Item, error)
	List() ([]models.Item, error)
	Get(itemID string) (*models.Item, error)
	Update(item *models.Item) (*models.Item, error)
	Delete(itemID string) error
}

// New create new client
// TODO refresh/ access token 가능 추가
func New(endpoint string, key string) Interface {
	client := &clientImpl{
		endpoint:  endpoint,
		sess:      request.NewSession(nil),
		keyHeader: "Bearer " + key,
	}

	client.todos = &todoImpl{client: client}

	return client
}

// Client todo item client
type clientImpl struct {
	endpoint  string
	sess      request.Interface
	keyHeader string

	todos *todoImpl
}

func (c *clientImpl) TodoService() TodoService {
	return c.todos
}

// todoImpl todo api
type todoImpl struct {
	client *clientImpl
}

func (t *todoImpl) Create(item *models.Item) (*models.Item, error) {
	resp, err := t.client.sess.Post(t.client.endpoint).
		Header(echo.HeaderAuthorization, t.client.keyHeader).
		JSON(item).Do()
	if err != nil {
		return nil, errors.Wrapf(err, "create")
	}

	if !resp.Success() {
		return nil, errors.New(resp.String())
	}

	var created models.Item

	defer resp.Body.Close()
	if err := resp.JSON(&created); err != nil {
		return nil, errors.Wrapf(err, "create")
	}

	return &created, nil
}

// List list todo item
func (t *todoImpl) List() ([]models.Item, error) {
	resp, err := t.client.sess.Get("%s", t.client.endpoint).
		Header(echo.HeaderAuthorization, t.client.keyHeader).
		Do()
	if err != nil {
		return nil, errors.Wrapf(err, "list")
	}

	if !resp.Success() {
		return nil, errors.New(resp.String())
	}

	items := make([]models.Item, 0)
	defer resp.Body.Close()
	if err := resp.JSON(&items); err != nil {
		return nil, errors.Wrapf(err, "list")
	}

	return items, nil
}

// Get get todo item
func (t *todoImpl) Get(itemID string) (*models.Item, error) {
	resp, err := t.client.sess.Get("%s/%s", t.client.endpoint, itemID).
		Header(echo.HeaderAuthorization, t.client.keyHeader).
		Do()
	if err != nil {
		return nil, errors.Wrapf(err, "get")
	}

	if !resp.Success() {
		return nil, errors.New(resp.String())
	}

	var item models.Item
	defer resp.Body.Close()
	if err := resp.JSON(&item); err != nil {
		return nil, errors.Wrapf(err, "get")
	}

	return &item, nil
}

// Update update todo item
func (t *todoImpl) Update(item *models.Item) (*models.Item, error) {
	resp, err := t.client.sess.Put("%s/%s", t.client.endpoint, item.ID).
		Header(echo.HeaderAuthorization, t.client.keyHeader).
		JSON(item).
		Do()
	if err != nil {
		return nil, errors.Wrapf(err, "update")
	}

	if !resp.Success() {
		return nil, errors.New(resp.String())
	}

	var updated models.Item
	defer resp.Body.Close()
	if err := resp.JSON(&updated); err != nil {
		return nil, errors.Wrapf(err, "update")
	}

	return &updated, nil
}

// Delete delete todo item
func (t *todoImpl) Delete(itemID string) error {
	resp, err := t.client.sess.Delete("%s/%s", t.client.endpoint, itemID).
		Header(echo.HeaderAuthorization, t.client.keyHeader).
		Do()
	if err != nil {
		return errors.Wrapf(err, "delete")
	}

	if !resp.Success() {
		return errors.New(resp.String())
	}

	return nil
}
