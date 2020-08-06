package todo

//go:generate swag init -g app.go
import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	_ "github.com/whitekid/go-todo/pkg/docs" // swagger docs
	"github.com/whitekid/go-todo/pkg/models"
	"github.com/whitekid/go-todo/pkg/storage"
	session_storage "github.com/whitekid/go-todo/pkg/storage/session"
	log "github.com/whitekid/go-utils/logging"
	"github.com/whitekid/go-utils/service"
)

// New create new todo service
func New() service.Interface {
	return &todoService{}
}

// HTTPError type alias for workaround swagger schema
type HTTPError = echo.HTTPError

// Context echo custom Context
type Context struct {
	echo.Context
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
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("todo-secret"))))
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &Context{c}
			return next(cc)
		}
	})

	e.POST("/", s.HandleItemCreate)
	e.GET("/", s.handleItemList)
	e.GET("/:item_id", s.handleItemGet)
	e.PUT("/:item_id", s.handleItemUpdate)
	e.DELETE("/:item_id", s.handleItemDelete)

	newGoogleOAuthHandler().Route(e.Group("/auth"))

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	return e
}

// @summary list todo item
// @description list todo item
// @tags todo
// @success 200 {array} models.Item
// @router / [get]
func (s *todoService) handleItemList(c echo.Context) error {
	storage := session_storage.New(c)
	items, _ := storage.TodoService().List()
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

	stg := session_storage.New(c)
	item, err := stg.TodoService().Get(itemID)
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

	todos := session_storage.New(c).TodoService()
	if err := todos.Update(&item); err != nil {
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
func (s *todoService) handleItemDelete(c echo.Context) error {
	itemID := c.Param("item_id")
	if itemID == "" {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	if err := session_storage.New(c).TodoService().Delete(itemID); err != nil {
		if err == storage.ErrNotFound {
			return c.NoContent(http.StatusNoContent)
		}
		return err
	}

	return c.NoContent(http.StatusNoContent)
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

	if err := session_storage.New(c).TodoService().Create(&item); err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, &item)
}
