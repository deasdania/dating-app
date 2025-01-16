package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/deasdania/dating-app/config"
	"github.com/deasdania/dating-app/models"
	"github.com/deasdania/dating-app/status"
	smodels "github.com/deasdania/dating-app/storage/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Helper to log and return errors with a standard response
func (h *Handlers) RespondWithError(c echo.Context, statusCode int64, errCode status.DatingStatusCode, errMsg string) error {
	h.log.Errorf("Error [%s]: %s", errCode, errMsg) // Log the error with code and message
	c.JSON(int(statusCode), models.NewResponseError(statusCode, errCode, errMsg))
	return nil
}

// ExtractUserIDFromToken extracts the user ID from the Authorization token
func (h *Handlers) ExtractUserIDFromToken(c echo.Context) (*uuid.UUID, error) {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return nil, errors.New("missing authorization header")
	}

	authHeaderParts := strings.Split(authHeader, " ")
	if len(authHeaderParts) != 2 || authHeaderParts[0] != "Bearer" {
		return nil, errors.New("invalid authorization header format")
	}

	token := authHeaderParts[1]
	userID, err := config.ExtractIDFromToken(token, h.secret)
	if err != nil {
		return nil, fmt.Errorf("failed to extract user ID: %v", err)
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %v", err)
	}

	return &uid, nil
}

func (h *Handlers) GetProfile(c echo.Context) error {
	// Extract user ID from token
	uid, err := h.ExtractUserIDFromToken(c)
	if err != nil {
		h.log.Errorf("Failed to extract user ID from token: %v", err) // Log the error
		return h.RespondWithError(c, http.StatusUnauthorized, status.UserErrCode_Unauthorized, err.Error())
	}

	// Get profile from core
	ctx := c.Request().Context()
	profile, err := h.core.GetProfile(ctx, uid)
	if err != nil {
		h.log.Errorf("Error fetching profile for user ID %v: %v", uid, err) // Log the error with context
		return h.RespondWithError(c, http.StatusInternalServerError, status.SystemErrCode_Generic, err.Error())
	}

	c.JSON(http.StatusOK, gin.H{"profile": profile})
	return nil
}

func (h *Handlers) SetProfile(c echo.Context) error {
	var profile smodels.Profile
	if err := c.Bind(&profile); err != nil {
		h.log.Errorf("Failed to bind profile data: %v", err) // Log binding error
		return h.RespondWithError(c, http.StatusBadRequest, status.UserErrCode_InvalidRequest, err.Error())
	}

	if err := validateStruct(h.validate, profile); err != nil {
		h.log.Errorf("Validation error for profile: %v", err) // Log validation error
		return h.RespondWithError(c, http.StatusBadRequest, status.UserErrCode_InvalidRequest, err.Error())
	}

	// Extract user ID from token
	uid, err := h.ExtractUserIDFromToken(c)
	if err != nil {
		h.log.Errorf("Failed to extract user ID from token: %v", err) // Log the error
		return h.RespondWithError(c, http.StatusUnauthorized, status.UserErrCode_Unauthorized, err.Error())
	}

	// Set profile in core
	ctx := c.Request().Context()
	err = h.core.SetProfile(ctx, uid, &profile)
	if err != nil {
		h.log.Errorf("Error setting profile for user ID %v: %v", uid, err) // Log the error with context
		return h.RespondWithError(c, http.StatusInternalServerError, status.SystemErrCode_Generic, err.Error())
	}

	c.JSON(http.StatusOK, gin.H{"message": "successfully update the profile"})
	return nil
}

func (h *Handlers) GetPeopleProfiles(c echo.Context) error {
	// Extract user ID from token (optional in this case, but you can validate it)
	uid, err := h.ExtractUserIDFromToken(c)
	if err != nil {
		h.log.Errorf("Failed to extract user ID from token: %v", err) // Log the error
		return h.RespondWithError(c, http.StatusUnauthorized, status.UserErrCode_Unauthorized, err.Error())
	}

	// Pagination logic
	pageStr := c.QueryParam("page")
	limitStr := c.QueryParam("limit")

	// Default values
	page := uint(1)   // Default to 1 if not specified
	limit := uint(10) // Default to 10 if not specified

	// Convert page query param to uint
	if pageStr != "" {
		parsedPage, err := strconv.ParseUint(pageStr, 10, 32) // Parse as uint
		if err != nil {
			h.log.Errorf("Invalid page value: %v", err) // Log error
			return h.RespondWithError(c, http.StatusBadRequest, status.UserErrCode_InvalidRequest, "Invalid page parameter")
		}
		page = uint(parsedPage) // Store the parsed value
	}

	// Convert limit query param to uint
	if limitStr != "" {
		parsedLimit, err := strconv.ParseUint(limitStr, 10, 32) // Parse as uint
		if err != nil {
			h.log.Errorf("Invalid limit value: %v", err) // Log error
			return h.RespondWithError(c, http.StatusBadRequest, status.UserErrCode_InvalidRequest, "Invalid limit parameter")
		}
		limit = uint(parsedLimit) // Store the parsed value
	}

	h.log.Infof("page:%s, %d", pageStr, page)
	h.log.Infof("limit:%s, %d", limitStr, limit)
	// Call core function to get profiles (with pagination)
	ctx := c.Request().Context()
	profiles, err := h.core.GetPeopleProfiles(ctx, uid, page, limit)
	if err != nil {
		h.log.Errorf("Error fetching profiles with page %d and limit %d: %v", page, limit, err) // Log the error with pagination context
		return h.RespondWithError(c, http.StatusInternalServerError, status.SystemErrCode_Generic, err.Error())
	}

	c.JSON(http.StatusOK, gin.H{"profiles": profiles})
	return nil
}

func (h *Handlers) GetPeopleProfileByID(c echo.Context) error {
	// Extract user ID from token
	_, err := h.ExtractUserIDFromToken(c)
	if err != nil {
		h.log.Errorf("Failed to extract user ID from token: %v", err) // Log the error
		return h.RespondWithError(c, http.StatusUnauthorized, status.UserErrCode_Unauthorized, err.Error())
	}

	// Parse the profile ID from the URL parameter
	profileID := c.Param("id")
	if profileID == "" {
		h.log.Error("Missing profile ID in URL") // Log the error
		return h.RespondWithError(c, http.StatusBadRequest, status.UserErrCode_InvalidRequest, "Profile ID is required")
	}

	profileUID, err := uuid.Parse(profileID)
	if err != nil {
		h.log.Errorf("Invalid profile ID format: %v", err) // Log the error
		return h.RespondWithError(c, http.StatusBadRequest, status.UserErrCode_InvalidRequest, "Invalid profile ID")
	}

	// Fetch profile by ID
	ctx := c.Request().Context()
	profile, err := h.core.GetPeopleProfileByID(ctx, &profileUID)
	if err != nil {
		if err == sql.ErrNoRows {
			h.log.Errorf("Profile not found for ID %v: %v", profileUID, err) // Log if profile is not found
			return h.RespondWithError(c, http.StatusNotFound, status.UserErrCode_ProfileNotFound, "Profile not found")
		}
		h.log.Errorf("Error fetching profile by ID %v: %v", profileUID, err) // Log the error with profile ID context
		return h.RespondWithError(c, http.StatusInternalServerError, status.SystemErrCode_Generic, err.Error())
	}

	c.JSON(http.StatusOK, gin.H{"profile": profile})
	return nil
}
