package todo

//go:generate swag init -g app.go
import (
	"context"
	"log"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	"github.com/whitekid/go-todo/config"
	_ "github.com/whitekid/go-todo/docs" // swagger docs
	"github.com/whitekid/go-todo/handlers/auth"
	"github.com/whitekid/go-todo/handlers/oauth"
	"github.com/whitekid/go-todo/handlers/todo"
	"github.com/whitekid/go-todo/storage"
	. "github.com/whitekid/go-todo/types"
	"github.com/whitekid/go-utils/service"
)

// New create new todo service
func New() service.Interface {
	storage, err := storage.New("todo")
	if err != nil {
		panic(err)
	}

	return &todoService{
		storage: storage,
	}
}

// HTTPError type alias for workaround swagger schema
type HTTPError = echo.HTTPError

type todoService struct {
	storage storage.Interface
}

// @title TODO API
// @version 1.0
// @description This is a simple todo API service.
// @host
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func (s *todoService) Serve(ctx context.Context, args ...string) (err error) {
	e := s.setupRoute()

	defer s.storage.Close()

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		if err := e.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}()

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
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &Context{c}
			return next(cc)
		}
	})

	todo.New(s.storage).Route(e.Group(""))
	auth.New(s.storage).Route(e.Group("/auth"))
	oauth.New(s.storage, oauth.Options{
		ClientID:     config.ClientID(),
		ClientSecret: config.ClientSecret(),
		RedirectURL:  config.RootURL() + config.CallbackURL(),
	}).Route(e.Group("/oauth"))

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	return e
}
