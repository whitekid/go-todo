//Package oauth supports auth with google
package oauth

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	. "github.com/whitekid/go-todo/pkg/handlers/types"
	. "github.com/whitekid/go-todo/pkg/types"
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

func init() {
	rand.Seed(time.Now().UnixNano())
}

// New return google oauth handler
func New(opts Options) Handler {
	return &googleOAuthHandler{
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

	path string
}

func (g *googleOAuthHandler) Route(r Router) {
	r.GET("/", g.handleAuth)
	r.GET("/callback", g.handleCallback)
}

var charset = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// TODO move to go-utils
func randomString(l int) string {
	return strings.Map(func(r rune) rune { return charset[rand.Intn(len(charset))] }, string(make([]rune, l)))
}

func (g *googleOAuthHandler) oauthSession(c echo.Context) *sessions.Session {
	return c.(*Context).OauthSession()
}

func (g *googleOAuthHandler) session(c echo.Context) *sessions.Session {
	return c.(*Context).Session()
}

func (g *googleOAuthHandler) handleAuth(c echo.Context) error {
	session := g.session(c)
	value, ok := session.Values["email"]
	if !ok {
		return g.authenticate(c)
	}

	email, ok := value.(string)
	if !ok {
		return g.authenticate(c)
	}

	return c.String(http.StatusOK, email)
}

func (g *googleOAuthHandler) authenticate(c echo.Context) error {
	session := g.oauthSession(c)

	state := randomString(32)
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
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	// request access token
	client := g.oauthConf.Client(oauth2.NoContext, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	var user struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	sess = g.session(c)
	sess.Values["email"] = user.Email
	sess.Save(c.Request(), c.Response())

	return c.Redirect(http.StatusFound, "/")
}
