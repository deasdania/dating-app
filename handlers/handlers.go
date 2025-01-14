package handlers

import (
	"encoding/json"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/ory/viper"
	"github.com/sirupsen/logrus"

	"github.com/deasdania/dating-app/config"
	"github.com/deasdania/dating-app/core"
	ps "github.com/deasdania/dating-app/storage/postgresql"
)

// type Handlers struct {
// 	App      *echo.Echo
// 	Log      *logrus.Entry
// 	Validate *validator.Validate
// }

// type middlewareManager struct {
// 	jwtAuthM config.JWTAuthMiddleware
// }

// func NewHandlers(e *echo.Echo, mm *middlewareManager, timeoutContext time.Duration) *Handlers {

// 	data, err := json.MarshalIndent(e.Routes(), "", "  ")
// 	if err == nil {
// 		os.WriteFile("routes.json", data, 0o644)
// 	}
// 	return &Handlers{}
// }

// func (h *Handlers) signUp(c *gin.Context)          {}
// func (h *Handlers) login(c *gin.Context)           {}
// func (h *Handlers) getProfile(c *gin.Context)      {}
// func (h *Handlers) swipe(c *gin.Context)           {}
// func (h *Handlers) purchasePremium(c *gin.Context) {}
// func (h *Handlers) updateProfile(c *gin.Context)   {}

type API struct {
	App      *echo.Echo
	Log      *logrus.Entry
	Validate *validator.Validate
	Config   *viper.Viper
	Storage  *ps.Storage
}

type middlewareManager struct {
	jwtAuthM config.JWTAuthMiddleware
}

func (b *API) v1(e *echo.Echo, um *core.Core, mm *middlewareManager) {
	v1Group := e.Group("/v1", mm.jwtAuthM.JWTAuthMiddleware())
	v1GroupNoAuth := e.Group("/v1")

	data, err := json.MarshalIndent(e.Routes(), "", "  ")
	if err == nil {
		os.WriteFile("routes.json", data, 0o644)
	}
}

func Bootstrap(api *API) {
	timeoutContext := time.Duration(api.Config.GetInt("context.timeout")) * time.Second
	coreAPI := core.NewCore(
		api.Log,
		api.Storage,
		timeoutContext,
	)
	jwtAuthMiddleware := config.InitJWTAuthMiddleware(api.Config.GetString("access_token.secret"))

	mm := &middlewareManager{
		jwtAuthM: *jwtAuthMiddleware,
	}

	api.v1(api.App, coreAPI, mm)
}
