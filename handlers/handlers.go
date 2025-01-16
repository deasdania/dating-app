package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/deasdania/dating-app/config"
	"github.com/deasdania/dating-app/core"
	"github.com/deasdania/dating-app/status"
	"github.com/deasdania/dating-app/storage/models"
	ps "github.com/deasdania/dating-app/storage/postgresql"
	redis "github.com/deasdania/dating-app/storage/redis"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Handlers struct {
	app      *echo.Echo
	log      *logrus.Entry
	validate *validator.Validate
	config   *viper.Viper
	core     CoreI
	secret   string
}

func NewHandlers(
	app *echo.Echo, log *logrus.Entry, secret string, v1GroupNoAuth *echo.Group, v1GroupAuth *echo.Group, validate *validator.Validate, config *viper.Viper, core CoreI) {
	handler := &Handlers{
		app:      app,
		log:      log,
		validate: validate,
		config:   config,
		core:     core,
		secret:   secret,
	}
	v1GroupNoAuth.POST("/signup", handler.SignUp)
	v1GroupNoAuth.POST("/login", handler.Login)

	v1GroupAuth.POST("/profile", handler.SetProfile)
	v1GroupAuth.GET("/profile", handler.GetProfile)
	v1GroupAuth.GET("/profiles", handler.GetPeopleProfiles)
	v1GroupAuth.GET("/profiles/:id", handler.GetPeopleProfileByID)

	v1GroupAuth.POST("/swipe", handler.Swipe)
	v1GroupAuth.POST("/premium", handler.Premium)

}

type API struct {
	App      *echo.Echo
	Log      *logrus.Entry
	Validate *validator.Validate
	Config   *viper.Viper
	Storage  *ps.Storage
	RC       *redis.RedisConnection
}

type middlewareManager struct {
	jwtAuthM config.JWTAuthMiddleware
}

type CoreI interface {
	SignUp(ctx context.Context, user *models.User) error
	Login(ctx context.Context, input *models.User) (string, error)

	GetProfile(ctx context.Context, userID *uuid.UUID) (*models.Profile, error)
	SetProfile(ctx context.Context, userID *uuid.UUID, profile *models.Profile) error
	GetPeopleProfiles(ctx context.Context, userID *uuid.UUID, page, limit uint) ([]*models.Profile, error)
	GetPeopleProfileByID(ctx context.Context, profileID *uuid.UUID) (*models.Profile, error)

	Swipe(ctx context.Context, req *models.Swipe) (status.DatingStatusCode, error)
	SetPremium(ctx context.Context, userID *uuid.UUID, typeStr string) (status.DatingStatusCode, error)
}

func (b *API) v1(e *echo.Echo, um *core.Core, mm *middlewareManager) {
	v1Group := e.Group("/v1", mm.jwtAuthM.JWTAuthMiddleware())
	v1GroupNoAuth := e.Group("/v1")
	secret := b.Config.GetString("access_token.secret")

	NewHandlers(b.App, b.Log, secret, v1GroupNoAuth, v1Group, b.Validate, b.Config, um)

	data, err := json.MarshalIndent(e.Routes(), "", "  ")
	if err == nil {
		os.WriteFile("routes.json", data, 0o644)
	}
}

func Bootstrap(api *API) {
	timeoutContext := time.Duration(api.Config.GetInt("context.timeout")) * time.Second
	secret := api.Config.GetString("access_token.secret")
	coreAPI := core.NewCore(
		api.Log,
		api.Storage,
		api.RC,
		timeoutContext,
		secret,
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
