package auth

import (
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"github.com/whitekid/go-todo/pkg/config"
	"github.com/whitekid/go-todo/pkg/storage"
	"github.com/whitekid/go-todo/pkg/tokens"
	"github.com/whitekid/go-utils/request"
)

func TestAuth(t *testing.T) {
	e := echo.New()

	storage, err := storage.New("testdb") // TODO mock storage object
	require.NoError(t, err)
	defer storage.Close()
	handler := New(storage)
	handler.Route(e)

	ts := httptest.NewServer(e)
	defer ts.Close()

	email := "someone@here.com"
	token, err := tokens.New(email, config.RefreshTokenDuration())
	require.NoError(t, err)

	require.NoError(t, storage.TokenService().Create(email, token))

	resp, err := request.Put("%s/tokens", ts.URL).Header(echo.HeaderAuthorization, "Bearer "+token).Do()
	require.NoError(t, err)
	require.True(t, resp.Success(), "failed with status %d", resp.StatusCode)
	require.Equal(t, token, resp.Header.Get(echo.HeaderAuthorization))
}
