//Package oauth supports auth with google
package oauth

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/whitekid/go-todo/config"
	. "github.com/whitekid/go-todo/handlers/types"
	"github.com/whitekid/go-todo/storage"
	"github.com/whitekid/go-todo/tokens"
	. "github.com/whitekid/go-todo/types"
	"github.com/whitekid/go-utils"
	"github.com/whitekid/go-utils/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Options google oauth handler options
type Options struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Path         string
}

// New return google oauth handler
func New(storage storage.Interface, opts Options) Handler {
	return &googleOAuthHandler{
		storage: storage,
		oauthConf: &oauth2.Config{
			ClientID:     opts.ClientID,
			ClientSecret: opts.ClientSecret,
			RedirectURL:  opts.RedirectURL,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		},
		path: opts.Path,
	}
}

type googleOAuthHandler struct {
	oauthConf *oauth2.Config
	storage   storage.Interface

	path string
}

func (g *googleOAuthHandler) Route(r Router) {
	r.Use(
		session.Middleware(sessions.NewCookieStore([]byte("todo"))),
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
		})
	r.GET("/", g.handleAuth)
	r.GET("/callback", g.handleCallback)
}

func (g *googleOAuthHandler) oauthSession(c echo.Context) *sessions.Session {
	return c.(*Context).Get("oauth-session").(*sessions.Session)
}

func (g *googleOAuthHandler) handleAuth(c echo.Context) error {
	return g.authenticate(c)
}

func (g *googleOAuthHandler) authenticate(c echo.Context) error {
	session := g.oauthSession(c)

	state := utils.RandomString(32)
	session.Values["state"] = state
	session.Save(c.Request(), c.Response())

	return c.Redirect(http.StatusFound, g.oauthConf.AuthCodeURL(state))
}

func (g *googleOAuthHandler) handleCallback(c echo.Context) error {
	// check state is valid
	sess := g.oauthSession(c)
	state := sess.Values["state"]

	delete(sess.Values, "state")
	sess.Save(c.Request(), c.Response())

	if state != c.QueryParam("state") {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid state %v %v", state, c.QueryParam("state")))
	}

	// convert code to token
	token, err := g.oauthConf.Exchange(oauth2.NoContext, c.QueryParam("code"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// request access token
	client := g.oauthConf.Client(oauth2.NoContext, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	var user struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	refreshToken, err := tokens.New(user.Email, config.RefreshTokenDuration())
	if err != nil {
		log.Errorf("fail to generate refresh token: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	err = g.storage.TokenService().Create(user.Email, refreshToken)
	if err != nil {
		log.Errorf("fail to create refresh token: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	accessToken, err := tokens.New(user.Email, config.AccessTokenDuration())
	if err != nil {
		log.Errorf("fail to generate access token: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"refresh_token": refreshToken,
		"access_token":  accessToken,
	})
}
