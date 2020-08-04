package client

import (
	"github.com/pkg/errors"
	"github.com/whitekid/go-todo/pkg/models"
	"github.com/whitekid/go-utils/request"
)

// New create new client
func New(endpoint string) *Client {
	client := &Client{
		endpoint: endpoint,
		sess:     request.NewSession(nil),
	}

	client.Todos = &Todos{client: client}

	return client
}

// Client todo item client
type Client struct {
	endpoint string
	sess     request.Interface

	Todos *Todos
}

// Todos todo api
type Todos struct {
	client *Client
}

func (t *Todos) Create(item *models.Item) (*models.Item, error) {
	resp, err := t.client.sess.Post(t.client.endpoint).JSON(item).Do()
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
func (t *Todos) List() ([]models.Item, error) {
	resp, err := t.client.sess.Get("%s", t.client.endpoint).Do()
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
func (t *Todos) Get(itemID string) (*models.Item, error) {
	resp, err := t.client.sess.Get("%s/%s", t.client.endpoint, itemID).Do()
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
func (t *Todos) Update(item *models.Item) (*models.Item, error) {
	resp, err := t.client.sess.Put("%s/%s", t.client.endpoint, item.ID).JSON(item).Do()
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
func (t *Todos) Delete(itemID string) error {
	resp, err := t.client.sess.Delete("%s/%s", t.client.endpoint, itemID).Do()
	if err != nil {
		return errors.Wrapf(err, "delete")
	}

	if !resp.Success() {
		return errors.New(resp.String())
	}

	return nil
}
