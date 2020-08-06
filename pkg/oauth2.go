package todo

// authenticate against google oauth2
//
//

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2/google"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

// Router route to handler
type Router interface {
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
}

func newGoogleOAuthHandler() *googleOAuthHandler {
	return &googleOAuthHandler{
		oauthConf: &oauth2.Config{
			ClientID:     os.Getenv("TODO_CLIENT_ID"),
			ClientSecret: os.Getenv("TODO_CLIENT_SECRET"),
			RedirectURL:  "http://127.0.0.1:9998/auth/callback",
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		},
	}
}

type googleOAuthHandler struct {
	oauthConf *oauth2.Config
}

func (g *googleOAuthHandler) Route(r Router) {
	r.GET("/", g.handleAuth)
	r.GET("/callback", g.handleCallback)
}

func randToken() string {
	return "state" // TODO randomize this
}

func authSession(c echo.Context) *sessions.Session {
	store := sessions.NewCookieStore([]byte("secret"))
	session, _ := store.Get(c.Request(), "session")
	session.Options = &sessions.Options{
		Path:   "/auth", // TODO make configurable
		MaxAge: 300,
	}

	return session
}

func (g *googleOAuthHandler) handleAuth(c echo.Context) error {
	session := authSession(c)

	state := randToken()
	session.Values["state"] = state
	session.Save(c.Request(), c.Response())

	url := g.oauthConf.AuthCodeURL("state")

	return c.Redirect(http.StatusFound, url)
}

func (g *googleOAuthHandler) handleCallback(c echo.Context) error {
	// check state is valid
	session := authSession(c)
	state := session.Values["state"]

	delete(session.Values, "state")
	session.Save(c.Request(), c.Response())

	if state != c.QueryParam("state") {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("state mismatch %v %v", state, c.QueryParam("state")))
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
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	session.Values["user"] = user.Name
	session.Values["email"] = user.Email
	session.Save(c.Request(), c.Response())

	return c.JSON(http.StatusOK, map[string]string{
		"user":  user.Name,
		"email": user.Email,
	})
}
