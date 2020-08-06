package todo

import (
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
)

// Context  custom Context
type Context struct {
	echo.Context
}

func (c *Context) Session() *sessions.Session {
	return c.Get("session").(*sessions.Session)
}

func (c *Context) OauthSession() *sessions.Session {
	return c.Get("oauth-session").(*sessions.Session)
}
