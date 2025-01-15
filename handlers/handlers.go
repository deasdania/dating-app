package handlers

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/deasdania/dating-app/config"
	"github.com/deasdania/dating-app/core"
	ps "github.com/deasdania/dating-app/storage/postgresql"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
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

type Handlers struct {
	app      *echo.Echo
	log      *logrus.Entry
	validate *validator.Validate
	config   *viper.Viper
	core     *core.Core
}

func NewHandlers(
	app *echo.Echo, log *logrus.Entry, v1GroupNoAuth *echo.Group, v1GroupAuth *echo.Group, validate *validator.Validate, config *viper.Viper, core *core.Core) {
	handler := &Handlers{
		app:      app,
		log:      log,
		validate: validate,
		config:   config,
		core:     core,
	}
	v1GroupNoAuth.POST("/signup", handler.SignUp)

}

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

	NewHandlers(b.App, b.Log, v1GroupNoAuth, v1Group, b.Validate, b.Config, um)

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

// validateStruct validates a struct using the validator package
func validateStruct(val *validator.Validate, req interface{}) error {
	if err := val.Struct(req); err != nil {
		var keys []string
		for _, err := range err.(validator.ValidationErrors) {
			keys = append(keys, convertToSnakeCase(err.Field()))
		}
		combined := strings.Join(keys, ", ")
		if len(keys) > 1 {
			combined += " are"
		} else {
			combined += " is"
		}
		combined += " required"
		return fmt.Errorf(combined)
	}
	return nil
}

func convertToSnakeCase(input string) string {
	var result strings.Builder
	var prev rune

	for _, curr := range input {
		if unicode.IsUpper(curr) && prev != 0 {
			result.WriteRune('_')
		}
		result.WriteRune(unicode.ToLower(curr))
		prev = curr
	}

	return result.String()
}
