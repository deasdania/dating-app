package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/deasdania/dating-app/config"
	"github.com/deasdania/dating-app/models"
	"github.com/deasdania/dating-app/status"
	smodels "github.com/deasdania/dating-app/storage/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (h *Handlers) SignUp(c echo.Context) error {
	var user smodels.User
	if err := c.Bind(&user); err != nil {
		c.JSON(http.StatusBadRequest, models.NewResponseError(http.StatusBadRequest, status.UserErrCode_InvalidRequest, err.Error()))
		return err
	}
	ctx := c.Request().Context()
	if err := validateStruct(h.validate, user); err != nil {
		h.log.Error("err validator:", err)
		c.JSON(http.StatusBadRequest, models.NewResponseError(http.StatusBadRequest, status.UserErrCode_InvalidRequest, err.Error()))
		return err
	}

	if err := h.core.SignUp(ctx, &user); err != nil {
		c.JSON(http.StatusBadRequest, models.NewResponseError(http.StatusInternalServerError, status.SystemErrCode_Generic, err.Error()))
		return err
	}

	c.JSON(http.StatusOK, gin.H{"message": "User created successfully"})
	return nil
}

// Login route
func (h *Handlers) Login(c echo.Context) error {
	var user smodels.User
	if err := c.Bind(&user); err != nil {
		c.JSON(http.StatusBadRequest, models.NewResponseError(http.StatusBadRequest, status.UserErrCode_InvalidRequest, err.Error()))
		return err
	}
	ctx := c.Request().Context()
	if err := validateStruct(h.validate, user); err != nil {
		h.log.Error("err validator:", err)
		c.JSON(http.StatusBadRequest, models.NewResponseError(http.StatusBadRequest, status.UserErrCode_InvalidRequest, err.Error()))
		return err
	}

	var token string
	token, err := h.core.Login(ctx, &user)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewResponseError(http.StatusInternalServerError, status.SystemErrCode_Generic, err.Error()))
		return err
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
	return nil
}

func (h *Handlers) GetProfile(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	authHeaderparts := strings.Split(authHeader, " ")
	token := authHeaderparts[1]

	// Extract User ID
	getUserID, err := config.ExtractIDFromToken(
		token,
		h.secret,
	)
	if err != nil {
		h.log.Error("extract user_id : ", err.Error())
		c.JSON(http.StatusBadRequest, models.NewResponseError(http.StatusUnauthorized, status.UserErrCode_Unauthorized, err.Error()))
		return err
	}
	ctx := c.Request().Context()
	uid, err := uuid.Parse(getUserID)

	profile, err := h.core.GetProfile(ctx, &uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewResponseError(http.StatusInternalServerError, status.SystemErrCode_Generic, err.Error()))
		return err
	}

	c.JSON(http.StatusOK, gin.H{"profile": profile})
	return nil
}

func (h *Handlers) SetProfile(c echo.Context) error {
	var profile smodels.Profile
	if err := c.Bind(&profile); err != nil {
		c.JSON(http.StatusBadRequest, models.NewResponseError(http.StatusBadRequest, status.UserErrCode_InvalidRequest, err.Error()))
		return err
	}
	ctx := c.Request().Context()
	if err := validateStruct(h.validate, profile); err != nil {
		h.log.Error("err validator:", err)
		c.JSON(http.StatusBadRequest, models.NewResponseError(http.StatusBadRequest, status.UserErrCode_InvalidRequest, err.Error()))
		return err
	}

	authHeader := c.Request().Header.Get("Authorization")
	authHeaderparts := strings.Split(authHeader, " ")
	token := authHeaderparts[1]

	// Extract User ID
	getUserID, err := config.ExtractIDFromToken(
		token,
		h.secret,
	)
	if err != nil {
		h.log.Error("extract user_id : ", err.Error())
		c.JSON(http.StatusBadRequest, models.NewResponseError(http.StatusUnauthorized, status.UserErrCode_Unauthorized, err.Error()))
		return err
	}
	uid, err := uuid.Parse(getUserID)

	isNew, err := h.core.SetProfile(ctx, &uid, &profile)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewResponseError(http.StatusInternalServerError, status.SystemErrCode_Generic, err.Error()))
		return err
	}

	msg := "update"
	if isNew {
		msg = "create"
	}
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("successfully %s the profile", msg)})
	return nil
}
