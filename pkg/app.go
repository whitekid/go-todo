package todo

//go:generate swag init -g app.go
import (
	"context"
	"os"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	_ "github.com/whitekid/go-todo/pkg/docs" // swagger docs
	"github.com/whitekid/go-todo/pkg/handlers/auth"
	"github.com/whitekid/go-todo/pkg/handlers/oauth"
	"github.com/whitekid/go-todo/pkg/handlers/todo"
	. "github.com/whitekid/go-todo/pkg/types"
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
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("todo"))),
		func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				sess, _ := session.Get("session", c)
				sess.Options = &sessions.Options{
					Path:   "/",
					MaxAge: 86400,
				}
				c.Set("session", sess)

				return next(c)
			}
		})

	todo.New().Route(e.Group(""))
	auth.New().Route(e.Group("/auth"))
	oauth.New(oauth.Options{
		ClientID:     os.Getenv("TODO_CLIENT_ID"),
		ClientSecret: os.Getenv("TODO_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("TODO_ROOT_URL") + "/oauth/callback", // TODO configurable redirectURL
	}).Route(e.Group("/oauth",
		func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				sess, _ := session.Get("oauth", c)
				sess.Options = &sessions.Options{
					Path:   "/oauth",
					MaxAge: 300,
				}
				c.Set("oauth-session", sess)

				return next(c)
			}
		}))

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	return e
}
