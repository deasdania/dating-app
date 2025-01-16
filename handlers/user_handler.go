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
func (h *Handlers) handleRequest(c echo.Context, user *smodels.User) error {
	// Bind the request data into the user struct
	if err := c.Bind(user); err != nil {
		c.JSON(http.StatusBadRequest, models.NewResponseError(http.StatusBadRequest, status.UserErrCode_InvalidRequest, err.Error()))
		return err
	}

	// Validate the struct
	if err := validateStruct(h.validate, *user); err != nil {
		h.log.Error("Validation error:", err)
		c.JSON(http.StatusBadRequest, models.NewResponseError(http.StatusBadRequest, status.UserErrCode_InvalidRequest, err.Error()))
		return err
	}

	return nil
}

// SignUp route
func (h *Handlers) SignUp(c echo.Context) error {
	var user smodels.User

	// Use the helper function to handle binding and validation
	if err := h.handleRequest(c, &user); err != nil {
		return err // The error response is already handled in the helper function
	}

	// Call core.SignUp to register the user
	ctx := c.Request().Context()
	if err := h.core.SignUp(ctx, &user); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewResponseError(http.StatusInternalServerError, status.SystemErrCode_Generic, err.Error()))
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
	if err := h.handleRequest(c, &user); err != nil {
		return err // The error response is already handled in the helper function
	}

	// Call core.Login to log the user in and get the token
	ctx := c.Request().Context()
	token, err := h.core.Login(ctx, &user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewResponseError(http.StatusInternalServerError, status.SystemErrCode_Generic, err.Error()))
		return err
	}

	// Respond with the token
	c.JSON(http.StatusOK, gin.H{"token": token})
	return nil
}
