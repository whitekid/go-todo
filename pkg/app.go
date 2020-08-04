package todo

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
	log "github.com/whitekid/go-utils/logging"
	"github.com/whitekid/go-utils/service"
)

// New create new todo service
func New() service.Interface {
	return &todoService{}
}

type todoService struct {
}

func (s *todoService) Serve(ctx context.Context, args ...string) error {
	e := s.setupRoute()
	return e.Start("127.0.0.1:9998")
}

func (s *todoService) setupRoute() *echo.Echo {
	e := echo.New()

	loggerConfig := middleware.DefaultLoggerConfig
	e.Use(middleware.LoggerWithConfig(loggerConfig))
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("todo-secret"))))

	e.POST("/", s.handleItemCreate)
	e.GET("/", s.handleItemList)
	e.GET("/:item_id", s.handleItemGet)
	e.PUT("/:item_id", s.handleItemUpdate)
	e.DELETE("/:item_id", s.handleItemDelete)

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

func (s *todoService) items(c echo.Context) []todoItem {
	sess := s.session(c)

	itemsV, ok := sess.Values["items"]
	if !ok {
		itemsV = []byte{}
	}

	items := make([]todoItem, 0)
	buf, ok := itemsV.([]byte)
	b := bytes.NewBuffer(buf)
	if err := json.NewDecoder(b).Decode(&items); err != nil {
		log.Errorf("json decode failed: %s, buf: %s", err, string(buf))
	}

	return items
}

func (s *todoService) saveItems(items []todoItem, c echo.Context) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(items); err != nil {
		return errors.Wrapf(err, "saveItems")
	}

	sess := s.session(c)
	sess.Values["items"] = buf.Bytes()

	log.Infof("save items %+v, data: %s", items, buf.String())
	return sess.Save(c.Request(), c.Response())
}

func (s *todoService) handleItemList(c echo.Context) error {
	items := s.items(c)
	return c.JSON(http.StatusOK, items)
}

func (s *todoService) handleItemGet(c echo.Context) error {
	itemID := c.Param("item_id")
	if itemID == "" {
		return c.NoContent(http.StatusNotFound)
	}

	items := s.items(c)
	for _, item := range items {
		if item.ID == itemID {
			return c.JSON(http.StatusOK, item)
		}
	}
	return c.NoContent(http.StatusNotFound)
}

func (s *todoService) handleItemUpdate(c echo.Context) error {
	itemID := c.Param("item_id")
	if itemID == "" {
		return c.NoContent(http.StatusNotFound)
	}

	var update todoItem
	if err := c.Bind(&update); err != nil {
		return errors.Wrapf(err, "ItemUpdate")
	}

	if err := update.validate(); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	items := s.items(c)
	for i, item := range items {
		if item.ID == itemID {
			items[i] = update
			s.saveItems(items, c)
			return c.JSON(http.StatusAccepted, &items[i])
		}
	}
	return c.NoContent(http.StatusNotFound)
}

func (s *todoService) handleItemDelete(c echo.Context) error {
	itemID := c.Param("item_id")
	if itemID == "" {
		return c.NoContent(http.StatusNotFound)
	}

	items := s.items(c)
	for i, item := range items {
		if item.ID == itemID {
			items := append(items[:i], items[i+1:]...)
			s.saveItems(items, c)
			return c.JSON(http.StatusAccepted, item)
		}
	}
	return c.NoContent(http.StatusNotFound)
}

func (s *todoService) handleItemCreate(c echo.Context) error {
	var item todoItem
	if err := c.Bind(&item); err != nil {
		return errors.Wrapf(err, "IndexPost")
	}

	item.ID = uuid.New().String()
	if err := item.validate(); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	items := s.items(c)
	items = append(items, item)

	s.saveItems(items, c)

	return c.JSON(http.StatusCreated, item)
}

type todoItem struct {
	ID      string    `json:"id"`
	Title   string    `json:"title"`
	DueDate time.Time `json:"due_date"`
	Rank    int       `json:"rank"`
}

func (i *todoItem) validate() error {
	if i.Title == "" {
		return errors.New("title required")
	}

	return nil
}
