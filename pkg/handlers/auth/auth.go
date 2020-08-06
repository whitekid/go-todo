package auth

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	. "github.com/whitekid/go-todo/pkg/handlers/types"
	. "github.com/whitekid/go-todo/pkg/types"
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
	sess := c.(*Context).Session()
	value, ok := sess.Values["email"]
	if !ok {
		return c.String(http.StatusOK, "no email session key")
	}

	email, ok := value.(string)
	if !ok {
		return c.String(http.StatusOK, "invalid type")
	}

	return c.String(http.StatusOK, fmt.Sprintf("|%s|", email))
}

func (h *authHandler) handleLogout(c echo.Context) error {
	sess := c.(*Context).Session()

	delete(sess.Values, "email")
	sess.Save(c.Request(), c.Response())

	return c.Redirect(http.StatusFound, "/")
}
