package client

import (
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/whitekid/go-utils/log"
)

type authImpl struct {
	client *clientImpl
}

func (a *authImpl) refreshAccessToken() error {
	log.Debug("refresh access token")

	resp, err := a.client.sess.Put("%s/auth/tokens", a.client.endpoint).
		Header(echo.HeaderAuthorization, "Bearer "+a.client.refreshToken).
		Do()
	if err != nil {
		return err
	}

	if !resp.Success() {
		return errors.New(resp.String())
	}

	token := resp.Header.Get(echo.HeaderAuthorization)
	if token == "" {
		log.Errorf("response header empty")
		return errors.New("invalid response")
	}

	a.client.accessToken = token
	return nil
}
