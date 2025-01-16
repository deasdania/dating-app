package handlers

import (
	"net/http"

	"github.com/deasdania/dating-app/status"
	"github.com/deasdania/dating-app/storage/models"
	"github.com/gin-gonic/gin"
	"github.com/labstack/echo/v4"
)

func (h *Handlers) UpdatePremiumStatus(c echo.Context) error {
	// Extract user ID from token (optional in this case, but you can validate it)
	uid, err := h.ExtractUserIDFromToken(c)
	if err != nil {
		h.log.Errorf("Failed to extract user ID from token: %v", err) // Log the error
		return h.RespondWithError(c, http.StatusUnauthorized, status.UserErrCode_Unauthorized, err.Error())
	}

	ts := c.QueryParam("type")
	var packageT models.PackageType
	switch ts {
	case "remove_quota":
		packageT = models.RemoveQuota
	case "verified_label":
		packageT = models.VerifiedLabel
	default:
		return h.RespondWithError(c, http.StatusBadRequest, status.UserErrCode_InvalidRequestPremiumPackage, "")
	}

	ctx := c.Request().Context()
	statusCode, err := h.core.SetPremium(ctx, uid, packageT)
	if err != nil {
		h.log.Errorf("Error set premium type %s for user %s: %v", ts, uid, err)
		return h.RespondWithError(c, http.StatusInternalServerError, statusCode, err.Error())
	}

	c.JSON(http.StatusOK, gin.H{"message": "successfully set up"})
	return nil
}
