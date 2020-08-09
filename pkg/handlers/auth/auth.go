package auth

import (
	"net/http"

	"github.com/labstack/echo/v4"
	. "github.com/whitekid/go-todo/pkg/handlers/types"
)

func New() Handler {
	return &authHandler{}
}

type authHandler struct {
}

func (h *authHandler) Route(r Router) {
	r.GET("/", h.handleIndex)
	r.GET("/logout", h.handleLogout)
}

func (h *authHandler) handleIndex(c echo.Context) error {
	return c.String(http.StatusOK, "")
}

func (h *authHandler) handleLogout(c echo.Context) error {
	return c.String(http.StatusOK, "")
}
