package tokens

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	storage_types "github.com/whitekid/go-todo/storage/types"
	"github.com/whitekid/go-utils/log"
)

// TokenMiddleware header에서 token을 가져오고, 이를 검증한다.
// Authorization: Bearer {token} 형태로 전송전송한다. token은 jwt 형태이고, 안에 issuer에 email이 들어있음
// refreshToken=true면 token이 있는지까지 검사한다.
//
// 401 token expired
// 403 기타 오류 토큰 오류
func TokenMiddleware(storage storage_types.Interface, isRefreshToken bool) echo.MiddlewareFunc {
	return middleware.KeyAuth(func(key string, c echo.Context) (bool, error) {
		key = strings.TrimSpace(key)
		if key == "" {
			return false, echo.NewHTTPError(http.StatusUnauthorized)
		}

		email, err := Parse(key)
		if err != nil {
			if IsExpired(err) {
				return false, echo.NewHTTPError(http.StatusUnauthorized, err.Error())
			}

			return false, echo.NewHTTPError(http.StatusForbidden, err.Error())
		}

		if isRefreshToken {
			// refresh token should be exists
			token, err := storage.TokenService().Get(key)
			if err != nil {
				return false, echo.NewHTTPError(http.StatusForbidden, err.Error())
			}

			if _, err := Parse(token); err != nil {
				if IsExpired(err) {
					return false, echo.NewHTTPError(http.StatusUnauthorized, err.Error())
				}
				return false, echo.NewHTTPError(http.StatusForbidden, err.Error())
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
