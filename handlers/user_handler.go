package handlers

import (
	"net/http"

	"github.com/deasdania/dating-app/models"
	"github.com/deasdania/dating-app/status"
	smodels "github.com/deasdania/dating-app/storage/models"
	"github.com/gin-gonic/gin"
	"github.com/labstack/echo/v4"
)

// Helper function to handle request binding, validation, and response
func (h *Handlers) handleRequestUser(c echo.Context, user *smodels.User) error {
	// Bind the request data into the user struct
	if err := c.Bind(user); err != nil {
		c.JSON(http.StatusBadRequest, models.NewResponseError(http.StatusBadRequest, status.UserErrCode_InvalidRequest, err.Error()))
		return err
	}

	// Validate the struct
	if err := validateStruct(h.Validate, *user); err != nil {
		h.Log.Error("Validation error:", err)
		c.JSON(http.StatusBadRequest, models.NewResponseError(http.StatusBadRequest, status.UserErrCode_InvalidRequest, err.Error()))
		return err
	}

	return nil
}

// SignUp route
func (h *Handlers) SignUp(c echo.Context) error {
	var user smodels.User

	// Use the helper function to handle binding and validation
	if err := h.handleRequestUser(c, &user); err != nil {
		return err // The error response is already handled in the helper function
	}

	// Call core.SignUp to register the user
	ctx := c.Request().Context()
	// validate username and password values
	if user.Username == "" || user.Password == "" || user.Email == "" {
		c.JSON(http.StatusBadRequest, models.NewResponseError(http.StatusBadRequest, status.UserErrCode_RequestUsernamePasswordEmail, ""))
		return nil
	}
	if st, err := h.Core.SignUp(ctx, &user); err != nil {
		if st == status.UserErrCode_EmailIsTaken || st == status.UserErrCode_UsernameIsTaken {
			c.JSON(http.StatusInternalServerError, models.NewResponseError(http.StatusBadRequest, st, err.Error()))
			return err
		}
		c.JSON(http.StatusInternalServerError, models.NewResponseError(http.StatusInternalServerError, st, err.Error()))
		return err
	}

	// Respond with success message
	c.JSON(http.StatusOK, gin.H{"message": "User created successfully"})
	return nil
}

// Login route
func (h *Handlers) Login(c echo.Context) error {
	var user smodels.User

	// Use the helper function to handle binding and validation
	if err := h.handleRequestUser(c, &user); err != nil {
		return err // The error response is already handled in the helper function
	}

	// validate username and password values
	if user.Username == "" || user.Password == "" {
		c.JSON(http.StatusBadRequest, models.NewResponseError(http.StatusBadRequest, status.UserErrCode_RequestUsernamePassword, ""))
		return nil
	}
	// Call core.Login to log the user in and get the token
	ctx := c.Request().Context()
	st, token, err := h.Core.Login(ctx, &user)
	if err != nil {
		if st == status.SystemErrCode_FailedCompareHashPassword {
			c.JSON(http.StatusBadRequest, models.NewResponseError(http.StatusBadRequest, status.SystemErrCode_FailedCompareHashPassword, ""))
			return err
		}
		c.JSON(http.StatusInternalServerError, models.NewResponseError(http.StatusInternalServerError, st, err.Error()))
		return err
	}

	// Respond with the token
	c.JSON(http.StatusOK, gin.H{"token": token})
	return nil
}
