package client

import (
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
func New(endpoint string, refreshToken string) Interface {
	client := &clientImpl{
		endpoint:     endpoint,
		sess:         request.NewSession(nil),
		refreshToken: refreshToken,
	}

	client.todos = &todoImpl{client: client}
	client.auth = &authImpl{client: client}

	return client
}

// Client todo item client
type clientImpl struct {
	endpoint     string
	sess         request.Interface
	refreshToken string
	accessToken  string

	todos *todoImpl
	auth  *authImpl
}

func (c *clientImpl) TodoService() TodoService {
	return c.todos
}

func (c *clientImpl) ensureAccessToken() error {
	if c.accessToken == "" {
		if err := c.auth.refreshAccessToken(); err != nil {
			return err
		}
	}

	return nil
}
