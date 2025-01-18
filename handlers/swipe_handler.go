package handlers

import (
	"fmt"
	"net/http"

	"github.com/deasdania/dating-app/models"
	"github.com/deasdania/dating-app/status"
	smodels "github.com/deasdania/dating-app/storage/models"
	"github.com/gin-gonic/gin"
	"github.com/labstack/echo/v4"
)

// Helper function to handle request binding, validation, and response
func (h *Handlers) handleRequestSwipe(c echo.Context, swipe *smodels.Swipe) error {
	// Bind the request data into the swipe struct
	if err := c.Bind(swipe); err != nil {
		c.JSON(http.StatusBadRequest, models.NewResponseError(http.StatusBadRequest, status.UserErrCode_InvalidRequest, err.Error()))
		return err
	}

	// Validate the struct
	if err := validateStruct(h.Validate, *swipe); err != nil {
		h.Log.Error("Validation error:", err)
		c.JSON(http.StatusBadRequest, models.NewResponseError(http.StatusBadRequest, status.UserErrCode_InvalidRequest, err.Error()))
		return err
	}
	if swipe.ProfileID == nil {
		c.JSON(http.StatusBadRequest, models.NewResponseError(http.StatusBadRequest, status.UserErrCode_InvalidRequestProfileIDRequired, ""))
		return nil
	}
	if swipe.Direction == "" || (swipe.Direction != "left" && swipe.Direction != "right") {
		c.JSON(http.StatusBadRequest, models.NewResponseError(http.StatusBadRequest, status.UserErrCode_InvalidRequestDirectionRequired, ""))
		return nil
	}
	switch swipe.Direction {
	case "left":
		swipe.Direction = "pass"
	case "right":
		swipe.Direction = "like"
	}

	return nil
}

func (h *Handlers) Swipe(c echo.Context) error {
	var swipe smodels.Swipe
	// Use the helper function to handle binding and validation
	if err := h.handleRequestSwipe(c, &swipe); err != nil {
		return err // The error response is already handled in the helper function
	}

	// Call core.Login to log the swipe in and get the token
	ctx := c.Request().Context()
	// Extract user ID from token
	uid, err := h.ExtractUserIDFromToken(c)
	if err != nil {
		h.Log.Error(fmt.Sprintf("Failed to extract user ID from token: %v", err)) // Log the error
		return h.RespondWithError(c, http.StatusUnauthorized, status.UserErrCode_Unauthorized, err.Error())
	}
	swipe.UserID = uid

	status, err := h.Core.Swipe(ctx, &swipe)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewResponseError(http.StatusInternalServerError, status, err.Error()))
		return err
	}

	c.JSON(http.StatusOK, gin.H{"message": "Swiped successfully"})
	return nil
}
