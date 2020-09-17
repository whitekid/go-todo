module github.com/whitekid/go-todo

go 1.15

replace github.com/dgraph-io/badger/v2 => github.com/dgraph-io/badger/v2 v2.0.3

require (
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751
	github.com/dgraph-io/badger/v2 v2.0.0-00010101000000-000000000000
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/go-playground/validator/v10 v10.4.1
	github.com/golang/mock v1.4.4
	github.com/google/uuid v1.1.2
	github.com/gorilla/sessions v1.2.1
	github.com/labstack/echo-contrib v0.9.0
	github.com/labstack/echo/v4 v4.1.17
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.6.1
	github.com/swaggo/echo-swagger v1.0.0
	github.com/swaggo/swag v1.6.9
	github.com/whitekid/go-utils v0.0.0-20201127140658-51e8514b0035
	golang.org/x/oauth2 v0.0.0-20201109201403-9fd604954f58
)
