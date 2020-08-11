package todo

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"github.com/whitekid/go-todo/storage"
	"github.com/whitekid/go-utils/request"
)

func TestList(t *testing.T) {
	storage, err := storage.New("testdb")
	require.NoError(t, err)
	handler := New(storage)
	e := echo.New()
	handler.Route(e)

	ts := httptest.NewServer(e)
	defer ts.Close()

	token, err := handler.(*todoHandler).storage.TokenService().Create("whitekid@gmail.com")
	require.NoError(t, err)

	resp, err := request.Get(ts.URL).Header(echo.HeaderAuthorization, fmt.Sprintf("Bearer %s", token.Token)).Do()
	require.NoError(t, err)
	require.Truef(t, resp.Success(), "code: %d", resp.StatusCode)
}
