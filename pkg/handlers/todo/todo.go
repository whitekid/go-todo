package todo

import (
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	. "github.com/whitekid/go-todo/pkg/handlers/types"
	"github.com/whitekid/go-todo/pkg/models"
	"github.com/whitekid/go-todo/pkg/storage"
	log "github.com/whitekid/go-utils/logging"
)

// New create todo handler
func New(storage storage.Interface) Handler {
	return &todoHandler{
		storage: storage,
	}
}

type todoHandler struct {
	storage storage.Interface
}

func (h *todoHandler) Route(r Router) {
	r.Use(middleware.KeyAuth(func(key string, c echo.Context) (bool, error) {
		key = strings.TrimSpace(key)
		if key == "" {
			return false, echo.NewHTTPError(http.StatusUnauthorized)
		}

		token, err := h.storage.TokenService().Get(key)
		if err != nil {
			return false, echo.NewHTTPError(http.StatusForbidden)
		}

		if token.Expire.Before(time.Now()) {
			log.Infof("token expired: %s", key)
			h.storage.TokenService().Delete(key)
			return false, echo.NewHTTPError(http.StatusBadRequest)
		}
		c.Set("token", token)

		user, err := h.storage.UserService().Get(token.Email)
		if err != nil {
			log.Error("token found, but user not found: %+v", token)
			return false, echo.NewHTTPError(http.StatusForbidden)
		}
		c.Set("user", user)

		return true, nil
	}))

	r.POST("/", h.handleCreate)
	r.GET("/", h.handleList)
	r.GET("/:item_id", h.handleGet)
	r.PUT("/:item_id", h.handleUpdate)
	r.DELETE("/:item_id", h.handleDelete)
}

func (h *todoHandler) token(c echo.Context) *storage.AccessToken {
	return c.Get("token").(*storage.AccessToken)
}

func (h *todoHandler) user(c echo.Context) *storage.User {
	return c.Get("user").(*storage.User)
}

// @summary create todo item
// @description do ping
// @tags todo
// @accept json
// @produce json
// @param todo body models.Item true "todo item"
// @success 201 {object} models.Item
// @failure 401 {object} HTTPError
// @failure 400 {object} HTTPError
// @failure 403 {object} HTTPError
// @router / [post]
// @Security ApiKeyAuth
func (h *todoHandler) handleCreate(c echo.Context) error {
	var item models.Item

	if err := c.Bind(&item); err != nil {
		log.Errorf("bind failed: %s", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	item.ID = uuid.New().String()
	if err := item.Validate(); err != nil {
		log.Errorf("validate failed: %s", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.storage.TodoService().Create(h.token(c).Email, &item); err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, &item)
}

// @summary list todo item
// @description list todo item
// @tags todo
// @success 200 {array} models.Item
// @failure 401 {object} HTTPError
// @failure 403 {object} HTTPError
// @router / [get]
// @Security ApiKeyAuth
func (h *todoHandler) handleList(c echo.Context) error {
	items, err := h.storage.TodoService().List(h.token(c).Email)
	if err != nil && err == storage.ErrNotAuthenticated {
		return echo.NewHTTPError(http.StatusForbidden, err)
	}
	return c.JSON(http.StatusOK, items)
}

// @summary get todo item
// @description get todo item
// @tags todo
// @param item_id path string true "todo item ID"
// @success 200 {object} models.Item
// @failure 401 {object} HTTPError
// @failure 403 {object} HTTPError
// @failure 404 {object} HTTPError
// @router /{item_id} [get]
// @Security ApiKeyAuth
func (h *todoHandler) handleGet(c echo.Context) error {
	itemID := c.Param("item_id")
	if itemID == "" {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	item, err := h.storage.TodoService().Get(h.token(c).Email, itemID)
	if err != nil {
		switch err {
		case storage.ErrNotFound:
			return echo.NewHTTPError(http.StatusNotFound)
		case storage.ErrNotAuthenticated:
			return echo.NewHTTPError(http.StatusForbidden, err)
		}
		return err
	}
	return c.JSON(http.StatusOK, item)
}

// @summary update todo item
// @description update todo item
// @tags todo
// @param item_id path string true "todo item ID"
// @param item body models.Item true "todo item"
// @success 202 {object} models.Item
// @failure 400 {object} HTTPError
// @failure 401 {object} HTTPError
// @failure 403 {object} HTTPError
// @failure 404 {object} HTTPError
// @router /{item_id} [put]
// @Security ApiKeyAuth
func (h *todoHandler) handleUpdate(c echo.Context) error {
	itemID := c.Param("item_id")
	if itemID == "" {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	var item models.Item
	if err := c.Bind(&item); err != nil {
		log.Errorf("ItemUpdate failed: %s", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if item.ID != itemID {
		return echo.NewHTTPError(http.StatusBadRequest, "item ID must be same to given to path")
	}

	if err := item.Validate(); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.storage.TodoService().Update(h.token(c).Email, &item); err != nil {
		switch err {
		case storage.ErrNotFound:
			return echo.NewHTTPError(http.StatusNotFound)
		case storage.ErrNotAuthenticated:
			return echo.NewHTTPError(http.StatusForbidden, err)
		}

		return err
	}

	return c.JSON(http.StatusAccepted, &item)
}

// @summary delete todo item
// @description delete todo item
// @tags todo
// @param item_id path string true "todo item ID"
// @success 204 {string} string
// @failure 401 {object} HTTPError
// @failure 403 {object} HTTPError
// @failure 404 {object} HTTPError
// @router /{item_id} [delete]
// @Security ApiKeyAuth
func (h *todoHandler) handleDelete(c echo.Context) error {
	itemID := c.Param("item_id")
	if itemID == "" {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	if err := h.storage.TodoService().Delete(h.token(c).Email, itemID); err != nil {
		switch err {
		case storage.ErrNotFound:
			return echo.NewHTTPError(http.StatusNotFound)
		case storage.ErrNotAuthenticated:
			return echo.NewHTTPError(http.StatusForbidden, err)
		}
		return err
	}

	return c.NoContent(http.StatusNoContent)
}
