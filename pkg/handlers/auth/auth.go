package auth

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/whitekid/go-todo/pkg/config"
	. "github.com/whitekid/go-todo/pkg/handlers/types"
	"github.com/whitekid/go-todo/pkg/storage"
	"github.com/whitekid/go-todo/pkg/tokens"
)

// New create new auth handler
func New(storage storage.Interface) Handler {
	return &authHandler{
		storage: storage,
	}
}

type authHandler struct {
	storage storage.Interface
}

func (h *authHandler) Route(r Router) {
	r.POST("/tokens", h.handleTokenRefresh, tokens.TokenMiddleware(h.storage, true))
}

// refresh access token from refresh token
// TODO write swagger spec
func (h *authHandler) handleTokenRefresh(c echo.Context) error {
	email := c.Get("user").(*storage.User).Email
	token, err := tokens.New(email, config.RefreshTokenDuration())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	if err := h.storage.TokenService().Create(email, token); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.String(http.StatusOK, token)
}
