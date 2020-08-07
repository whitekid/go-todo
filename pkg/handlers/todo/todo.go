package todo

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	. "github.com/whitekid/go-todo/pkg/handlers/types"
	"github.com/whitekid/go-todo/pkg/models"
	"github.com/whitekid/go-todo/pkg/storage"
	. "github.com/whitekid/go-todo/pkg/storage/types"
	log "github.com/whitekid/go-utils/logging"
)

// New create todo handler
func New() Handler {
	storage, err := storage.New("todo")
	if err != nil {
		panic(err)
	}

	return &todoHandler{
		storage: storage,
	}
}

type todoHandler struct {
	storage storage.Interface
}

func (h *todoHandler) Route(r Router) {
	r.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// setup context
			cc, ok := h.storage.(Contexter)
			if ok {
				cc.SetContext(c)
			}

			return next(c)
		}
	})

	r.POST("/", h.handleCreate)
	r.GET("/", h.handleList)
	r.GET("/:item_id", h.handleGet)
	r.PUT("/:item_id", h.handleUpdate)
	r.DELETE("/:item_id", h.handleDelete)
}

// @summary create todo item
// @description do ping
// @tags todo
// @accept json
// @produce json
// @param todo body models.Item true "todo item"
// @success 201 {object} models.Item
// @failure 400 {object} HTTPError
// @router / [post]
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

	if err := h.storage.TodoService().Create(&item); err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, &item)
}

// @summary list todo item
// @description list todo item
// @tags todo
// @success 200 {array} models.Item
// @router / [get]
func (h *todoHandler) handleList(c echo.Context) error {
	items, _ := h.storage.TodoService().List()
	return c.JSON(http.StatusOK, items)
}

// @summary get todo item
// @description get todo item
// @tags todo
// @param item_id path string true "todo item ID"
// @success 200 {object} models.Item
// @failure 404 {object} HTTPError
// @router /{item_id} [get]
func (h *todoHandler) handleGet(c echo.Context) error {
	itemID := c.Param("item_id")
	if itemID == "" {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	item, err := h.storage.TodoService().Get(itemID)
	if err == storage.ErrNotFound {
		return echo.NewHTTPError(http.StatusNotFound)
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
// @failure 404 {object} HTTPError
// @router /{item_id} [put]
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

	if err := h.storage.TodoService().Update(&item); err != nil {
		if err == storage.ErrNotFound {
			return echo.NewHTTPError(http.StatusNotFound, err)
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
// @failure 404 {object} HTTPError
// @router /{item_id} [delete]
func (h *todoHandler) handleDelete(c echo.Context) error {
	itemID := c.Param("item_id")
	if itemID == "" {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	if err := h.storage.TodoService().Delete(itemID); err != nil {
		if err == storage.ErrNotFound {
			return c.NoContent(http.StatusNoContent)
		}
		return err
	}

	return c.NoContent(http.StatusNoContent)
}
