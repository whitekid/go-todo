package todo

//go:generate swag init -g app.go
import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
	echoSwagger "github.com/swaggo/echo-swagger"
	_ "github.com/whitekid/go-todo/pkg/docs" // swagger docs
	"github.com/whitekid/go-todo/pkg/models"
	log "github.com/whitekid/go-utils/logging"
	"github.com/whitekid/go-utils/service"
)

// New create new todo service
func New() service.Interface {
	return &todoService{}
}

// HTTPError type alias for workaround swagger schema
type HTTPError = echo.HTTPError

type todoService struct {
}

// @title TODO API
// @version 1.0
// @description This is a simple todo API service.
// @host
// @BasePath /
func (s *todoService) Serve(ctx context.Context, args ...string) error {
	e := s.setupRoute()
	return e.Start("127.0.0.1:9998")
}

func (s *todoService) setupRoute() *echo.Echo {
	e := echo.New()

	loggerConfig := middleware.DefaultLoggerConfig
	e.Use(middleware.LoggerWithConfig(loggerConfig))
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("todo-secret"))))

	e.POST("/", s.HandleItemCreate)
	e.GET("/", s.handleItemList)
	e.GET("/:item_id", s.handleItemGet)
	e.PUT("/:item_id", s.handleItemUpdate)
	e.DELETE("/:item_id", s.handleItemDelete)

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	return e
}

func (s *todoService) session(c echo.Context) *sessions.Session {
	sess, _ := session.Get("todo-session", c)
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: true,
	}

	return sess
}

func (s *todoService) items(c echo.Context) []models.Item {
	sess := s.session(c)

	itemsV, ok := sess.Values["items"]
	if !ok {
		itemsV = []byte{}
	}

	items := make([]models.Item, 0)
	buf, ok := itemsV.([]byte)
	b := bytes.NewBuffer(buf)
	if err := json.NewDecoder(b).Decode(&items); err != nil {
		log.Warnf("json decode failed: %s, buf: %s, reset to empty items", err, string(buf))
	}

	return items
}

func (s *todoService) saveItems(items []models.Item, c echo.Context) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(items); err != nil {
		return errors.Wrapf(err, "saveItems")
	}

	sess := s.session(c)
	sess.Values["items"] = buf.Bytes()

	log.Infof("save items %+v, data: %s", items, buf.String())
	return sess.Save(c.Request(), c.Response())
}

// @summary list todo item
// @description list todo item
// @tags todo
// @success 200 {array} models.Item
// @router / [get]
func (s *todoService) handleItemList(c echo.Context) error {
	items := s.items(c)
	return c.JSON(http.StatusOK, items)
}

// @summary get todo item
// @description get todo item
// @tags todo
// @param item_id path string true "todo item ID"
// @success 200 {object} models.Item
// @failure 404 {object} HTTPError
// @router /{item_id} [get]
func (s *todoService) handleItemGet(c echo.Context) error {
	itemID := c.Param("item_id")
	if itemID == "" {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	items := s.items(c)
	for _, item := range items {
		if item.ID == itemID {
			return c.JSON(http.StatusOK, &item)
		}
	}
	return echo.NewHTTPError(http.StatusNotFound)
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
func (s *todoService) handleItemUpdate(c echo.Context) error {
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

	items := s.items(c)
	for i, e := range items {
		if e.ID == itemID {
			items[i] = item
			s.saveItems(items, c)
			return c.JSON(http.StatusAccepted, &items[i])
		}
	}
	return echo.NewHTTPError(http.StatusNotFound)
}

// @summary delete todo item
// @description delete todo item
// @tags todo
// @param item_id path string true "todo item ID"
// @success 204 {string} string
// @failure 404 {object} HTTPError
// @router /{item_id} [delete]
func (s *todoService) handleItemDelete(c echo.Context) error {
	itemID := c.Param("item_id")
	if itemID == "" {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	items := s.items(c)
	for i, item := range items {
		if item.ID == itemID {
			items := append(items[:i], items[i+1:]...)
			s.saveItems(items, c)
			return c.NoContent(http.StatusNoContent)
		}
	}
	return echo.NewHTTPError(http.StatusNotFound)
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
func (s *todoService) HandleItemCreate(c echo.Context) error {
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

	items := s.items(c)
	items = append(items, item)

	s.saveItems(items, c)

	return c.JSON(http.StatusCreated, &item)
}
