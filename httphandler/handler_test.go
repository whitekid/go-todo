package httphandler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"github.com/whitekid/go-utils/request"
)

type helloHandler struct{}

func newHelloHandler() Interface { return &helloHandler{} }
func (h *helloHandler) Route(r Router) {
	r.GET("", func(c echo.Context) error { return c.String(http.StatusOK, "hello") })
}

type worldHandler struct{}

func newWorldHandler() Interface { return &worldHandler{} }
func (h *worldHandler) Route(r Router) {
	r.GET("", func(c echo.Context) error { return c.String(http.StatusOK, "world") })
}

func TestHandler(t *testing.T) {
	e := echo.New()
	newHelloHandler().Route(e.Group("/hello"))
	newWorldHandler().Route(e.Group("/world"))

	ts := httptest.NewServer(e)
	defer ts.Close()

	{
		resp, err := request.Get(ts.URL + "/hello").Do()
		require.NoError(t, err)
		require.Equal(t, "hello", resp.String())
	}

	{
		resp, err := request.Get(ts.URL + "/world").Do()
		require.NoError(t, err)
		require.Equal(t, "world", resp.String())
	}
}
