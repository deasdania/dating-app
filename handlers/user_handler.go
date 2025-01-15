package handlers

import (
	"net/http"

	"github.com/deasdania/dating-app/models"
	"github.com/deasdania/dating-app/status"
	smodels "github.com/deasdania/dating-app/storage/models"
	"github.com/gin-gonic/gin"
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

// Get profile route
// func getProfile(c *gin.Context) {
// 	var profiles []models.Profile
// 	db.Limit(10).Find(&profiles)

// 	c.JSON(http.StatusOK, gin.H{"profiles": profiles})
// }
