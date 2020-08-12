package client

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/whitekid/go-todo/models"
)

type todoImpl struct {
	client *clientImpl
}

func (t *todoImpl) Create(item *models.Item) (*models.Item, error) { return t.doCreate(item, true) }
func (t *todoImpl) doCreate(item *models.Item, refresh bool) (*models.Item, error) {
	if err := t.client.ensureAccessToken(); err != nil {
		return nil, err
	}

	resp, err := t.client.sess.Post(t.client.endpoint).
		Header(echo.HeaderAuthorization, "Bearer "+t.client.accessToken).
		JSON(item).Do()
	if err != nil {
		return nil, errors.Wrapf(err, "create")
	}

	if !resp.Success() {
		if refresh && resp.StatusCode == http.StatusUnauthorized {
			if err := t.client.auth.refreshAccessToken(); err != nil {
				return nil, err
			}
			return t.doCreate(item, false)
		}

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
func (t *todoImpl) List() ([]models.Item, error) { return t.doList(true) }
func (t *todoImpl) doList(refresh bool) ([]models.Item, error) {
	if err := t.client.ensureAccessToken(); err != nil {
		return nil, err
	}

	resp, err := t.client.sess.Get("%s", t.client.endpoint).
		Header(echo.HeaderAuthorization, "Bearer "+t.client.accessToken).
		Do()
	if err != nil {
		return nil, errors.Wrapf(err, "list")
	}

	if !resp.Success() {
		if refresh && resp.StatusCode == http.StatusUnauthorized {
			if err := t.client.auth.refreshAccessToken(); err != nil {
				return nil, err
			}
			return t.doList(false)
		}

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
func (t *todoImpl) Get(itemID string) (*models.Item, error) { return t.doGet(itemID, true) }
func (t *todoImpl) doGet(itemID string, refresh bool) (*models.Item, error) {
	if err := t.client.ensureAccessToken(); err != nil {
		return nil, err
	}

	resp, err := t.client.sess.Get("%s/%s", t.client.endpoint, itemID).
		Header(echo.HeaderAuthorization, "Bearer "+t.client.accessToken).
		Do()
	if err != nil {
		return nil, errors.Wrapf(err, "get")
	}

	if !resp.Success() {
		if refresh && resp.StatusCode == http.StatusUnauthorized {
			if err := t.client.auth.refreshAccessToken(); err != nil {
				return nil, err
			}
			return t.doGet(itemID, false)
		}

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
func (t *todoImpl) Update(item *models.Item) (*models.Item, error) { return t.doUpdate(item, true) }
func (t *todoImpl) doUpdate(item *models.Item, refresh bool) (*models.Item, error) {
	if err := t.client.ensureAccessToken(); err != nil {
		return nil, err
	}

	resp, err := t.client.sess.Put("%s/%s", t.client.endpoint, item.ID).
		Header(echo.HeaderAuthorization, "Bearer "+t.client.accessToken).
		JSON(item).
		Do()
	if err != nil {
		return nil, errors.Wrapf(err, "update")
	}

	if !resp.Success() {
		if refresh && resp.StatusCode == http.StatusUnauthorized {
			if err := t.client.auth.refreshAccessToken(); err != nil {
				return nil, err
			}
			return t.doUpdate(item, false)
		}

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
func (t *todoImpl) Delete(itemID string) error { return t.doDelete(itemID, true) }
func (t *todoImpl) doDelete(itemID string, refresh bool) error {
	if err := t.client.ensureAccessToken(); err != nil {
		return err
	}

	resp, err := t.client.sess.Delete("%s/%s", t.client.endpoint, itemID).
		Header(echo.HeaderAuthorization, "Bearer "+t.client.accessToken).
		Do()
	if err != nil {
		return errors.Wrapf(err, "delete")
	}

	if !resp.Success() {
		if refresh && resp.StatusCode == http.StatusUnauthorized {
			if err := t.client.auth.refreshAccessToken(); err != nil {
				return err
			}
			return t.doDelete(itemID, false)
		}

		return errors.New(resp.String())
	}

	return nil
}
