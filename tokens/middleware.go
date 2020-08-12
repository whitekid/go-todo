package tokens

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	storage_types "github.com/whitekid/go-todo/storage/types"
	"github.com/whitekid/go-utils/log"
)

func TokenMiddleware(storage storage_types.Interface, isRefreshToken bool) echo.MiddlewareFunc {
	return middleware.KeyAuth(func(key string, c echo.Context) (bool, error) {
		key = strings.TrimSpace(key)
		if key == "" {
			return false, echo.NewHTTPError(http.StatusUnauthorized)
		}

		email, err := Parse(key)
		if err != nil {
			if _, ok := err.(*ValidationError); ok {
				return false, echo.NewHTTPError(http.StatusForbidden, err)
			}
			return false, echo.NewHTTPError(http.StatusForbidden, err)
		}

		if isRefreshToken {
			// refresh token should be exists
			token, err := storage.TokenService().Get(key)
			if err != nil {
				return false, echo.NewHTTPError(http.StatusForbidden, err)
			}

			if _, err := Parse(token); err != nil {
				return false, echo.NewHTTPError(http.StatusForbidden, err)
			}
		}

		// get user informations
		user, err := storage.UserService().Get(email)
		if err != nil {
			log.Error("token found, but user not found: %+v", key)
			return false, echo.NewHTTPError(http.StatusUnauthorized)
		}

		c.Set("user", user)

		return true, nil
	})
}
