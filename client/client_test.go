package client

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"github.com/whitekid/go-todo/config"
	"github.com/whitekid/go-todo/models"
	"github.com/whitekid/go-todo/tokens"
)

func TestRefresh(t *testing.T) {
	refreshToken, _ := tokens.New("hello", time.Minute)

	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		parts := strings.SplitN(c.Request().Header.Get(echo.HeaderAuthorization), "Bearer ", 2)
		if _, err := tokens.Parse(parts[1]); err != nil {
			if tokens.IsExpired(err) {
				return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
			}
			return echo.NewHTTPError(http.StatusForbidden, err.Error())
		}

		return c.JSON(http.StatusOK, []models.Item{})
	})
	e.PUT("/auth/tokens", func(c echo.Context) error {
		accessToken, _ := tokens.New("hello", config.AccessTokenDuration())
		c.Response().Header().Set(echo.HeaderAuthorization, accessToken)
		return c.NoContent(http.StatusOK)
	})
	ts := httptest.NewServer(e)
	defer ts.Close()

	client := New(ts.URL, refreshToken)
	_, err := client.TodoService().List()
	require.NoError(t, err, "expired token shout be refreshed")
}
