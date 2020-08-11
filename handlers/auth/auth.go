package auth

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/whitekid/go-todo/config"
	. "github.com/whitekid/go-todo/handlers/types"
	"github.com/whitekid/go-todo/storage"
	"github.com/whitekid/go-todo/tokens"
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
	r.PUT("/tokens", h.handleTokenRefresh, tokens.TokenMiddleware(h.storage, true))
}

// refresh access token from refresh token
// @summary refresh access token using refresh token
// @description refresh token can be obtain /oauth with google authentication
// @tags auth
// @success 200 "access token"
// @header 200 {string} Authorization "the new access token"
// @failure 401
// @failure 403
// @router / [put]
// @security ApiKeyAuth
func (h *authHandler) handleTokenRefresh(c echo.Context) error {
	email := c.Get("user").(*storage.User).Email
	token, err := tokens.New(email, config.RefreshTokenDuration())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	if err := h.storage.TokenService().Create(email, token); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	c.Response().Header().Set(echo.HeaderAuthorization, token)
	return c.NoContent(http.StatusOK)
}
