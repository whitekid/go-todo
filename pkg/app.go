package todo

//go:generate swag init -g app.go
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
	"github.com/swaggo/echo-swagger"
	_ "github.com/whitekid/go-todo/pkg/docs"
	log "github.com/whitekid/go-utils/logging"
	"github.com/whitekid/go-utils/service"
)

// New create new todo service
func New() service.Interface {
	return &todoService{}
}

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

// @summary list todo item
// @description list todo item
// @tags todo
// @success 200 {string} string
// @router / [get]
func (s *todoService) handleItemList(c echo.Context) error {
	items := s.items(c)
	return c.JSON(http.StatusOK, items)
}

// @summary get todo item
// @description get todo item
// @tags todo
// @param item_id path string true "todo item ID"
// @success 200 {string} string
// @failure 404 {string} string
// @router /{item_id} [get]
func (s *todoService) handleItemGet(c echo.Context) error {
	itemID := c.Param("item_id")
	if itemID == "" {
		return c.NoContent(http.StatusNotFound)
	}

	items := s.items(c)
	for _, item := range items {
		if item.ID == itemID {
			return c.JSON(http.StatusOK, &item)
		}
	}
	return c.NoContent(http.StatusNotFound)
}

// @summary update todo item
// @description update todo item
// @tags todo
// @param item_id path string true "todo item ID"
// @param item body todoItem true "todo item"
// @success 202 {string} string
// @failure 400 {string} string
// @failure 404 {string} string
// @router /{item_id} [put]
func (s *todoService) handleItemUpdate(c echo.Context) error {
	itemID := c.Param("item_id")
	if itemID == "" {
		return c.NoContent(http.StatusNotFound)
	}

	var update todoItem
	if err := c.Bind(&update); err != nil {
		log.Errorf("ItemUpdate failed: %s", err)
		return c.String(http.StatusBadRequest, err.Error())
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

// @summary delete todo item
// @description delete todo item
// @tags todo
// @param item_id path string true "todo item ID"
// @success 202 {string} string
// @failure 404 {string} string
// @router /{item_id} [delete]
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
			return c.JSON(http.StatusAccepted, &item)
		}
	}
	return c.NoContent(http.StatusNotFound)
}

// @summary create todo item
// @description do ping
// @tags todo
// @accept json
// @produce json
// @param todo body todoItem true "todo item"
// @success 201 {string} string
// @failure 400 {string} string
// @router / [post]
func (s *todoService) HandleItemCreate(c echo.Context) error {
	var item todoItem
	if err := c.Bind(&item); err != nil {
		log.Errorf("bind failed: %s", err)
		return c.String(http.StatusBadRequest, err.Error())
	}

	item.ID = uuid.New().String()
	if err := item.validate(); err != nil {
		log.Errorf("validate failed: %s", err)
		return c.String(http.StatusBadRequest, err.Error())
	}

	items := s.items(c)
	items = append(items, item)

	s.saveItems(items, c)

	return c.JSON(http.StatusCreated, &item)
}

// todoItem todo item
type todoItem struct {
	ID      string  `json:"id" example:"628b92ab-6d95-4fbe-b7c6-09cf5cd8941c" format:"uuid"`
	Title   string  `json:"title" example:"do something in future"`
	DueDate DueDate `json:"due_date" example:"2006-01-02" format:"date"`
	Rank    int     `json:"rank" example:"1" format:"int"`
}

type DueDate struct {
	time.Time
}

func Today() (d DueDate) {
	d.Time = time.Now().UTC().Truncate(time.Hour * 24)
	return
}

func (d *DueDate) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}

	d.Time = t
	return nil
}

func (d *DueDate) MarshalJSON() ([]byte, error) {
	s := d.Format("2006-01-02")
	return json.Marshal(s)
}

func (i *todoItem) validate() error {
	if i.Title == "" {
		return errors.New("title required")
	}

	return nil
}
