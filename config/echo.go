package config

import (
	"net/http"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/spf13/viper"

	"github.com/deasdania/dating-app/models"
	"github.com/deasdania/dating-app/status"
)

func NewEcho(config *viper.Viper, validate *BaseValidator) *echo.Echo {
	e := echo.New()
	customMiddleware := InitMiddleware()
	e.Use(echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
		AllowOrigins: config.GetStringSlice("allow_hosts"),
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	e.Use(customMiddleware.CORS)
	e.Validator = validate

	e.HTTPErrorHandler = func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		if err == echo.ErrUnauthorized {
			statusRes := status.ResponseFromCode(status.UserErrCode_Unauthorized)
			err = c.JSON(http.StatusUnauthorized, models.ResponseBase{
				Status:  http.StatusUnauthorized,
				Details: statusRes,
				Data:    nil,
			})
		} else {
			statusRes := status.ResponseFromCode(status.SystemErrCode_Generic)
			err = c.JSON(http.StatusUnauthorized, models.ResponseBase{
				Status:  http.StatusUnauthorized,
				Details: statusRes,
				Data:    nil,
			})
		}

		e.DefaultHTTPErrorHandler(err, c)
	}

	return e
}
