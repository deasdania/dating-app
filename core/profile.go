package core

import (
	"context"
	"errors"
	"time"

	"github.com/deasdania/dating-app/status"
	"github.com/deasdania/dating-app/storage/models"
	"github.com/google/uuid"
)

func (c *Core) GetProfile(ctx context.Context, userID *uuid.UUID) (status.DatingStatusCode, *models.Profile, error) {
	c.log.Info("Starting get profile core")

	profile, err := c.getProfileByUserID(ctx, userID)
	if err != nil {
		c.log.Error("Failed to retrieve profile", err)
		return status.UserErrCode_ProfileNotFound, nil, err
	}

	return status.Success_Generic, profile, nil
}

func (c *Core) SetProfile(ctx context.Context, userID *uuid.UUID, req *models.Profile) (status.DatingStatusCode, error) {
	c.log.Info("Starting set profile core")

	profile, err := c.getProfileByUserID(ctx, userID)
	if err != nil {
		c.log.Error("Failed to retrieve existing profile", err)
		return status.UserErrCode_ProfileNotFound, err
	}

	if req.ImageURL == "" {
		req.ImageURL = profile.ImageURL
	}

	if err := c.storage.UpdateProfilePartial(ctx, &models.Profile{
		ID:          profile.ID,
		Description: req.Description,
		ImageURL:    req.ImageURL,
	}); err != nil {
		c.log.Error("Failed to update profile in database", err)
		return status.SystemErrCode_FailedStoreData, err
	}

	return status.Success_Generic, nil
}

func (c *Core) GetPeopleProfiles(ctx context.Context, userID *uuid.UUID, page, limit uint) (status.DatingStatusCode, []*models.Profile, error) {
	c.log.Info("Starting get people profiles core")

	date := time.Now().Format("2006-01-02")

	profile, err := c.getProfileByUserID(ctx, userID)
	if err != nil {
		c.log.Error("Failed to retrieve profile for user", err)
		return status.UserErrCode_ProfileNotFound, nil, err
	}

	// Get profiles already swiped by the user
	_, swipeProfileIDs, err := c.storage.GetSwipes(ctx, models.SwipeFilterByUserID(userID), models.SwipeFilterByCreatedAtDate(date))
	if err != nil {
		c.log.Error("Failed to retrieve swipes", err)
		return status.SystemErrCode_FailedSwipeCount, nil, errors.New("failed to get swipes")
	}

	// Get profiles excluding the user and the swiped profiles
	profiles, err := c.storage.GetProfiles(
		ctx,
		models.ProfileFilterByPage(page),
		models.ProfileFilterByLimit(limit),
		models.ProfileFilterByExcludeProfileIDs(append([]*uuid.UUID{&profile.ID}, swipeProfileIDs...)),
	)
	if err != nil {
		c.log.Error("Failed to retrieve profiles from database", err)
		return status.UserErrCode_ProfileNotFound, nil, err
	}

	return status.Success_Generic, profiles, nil
}

func (c *Core) GetPeopleProfileByID(ctx context.Context, profileID *uuid.UUID) (status.DatingStatusCode, *models.Profile, error) {
	c.log.Info("Starting get profile by ID core")

	profile, err := c.getProfileByID(ctx, profileID)
	if err != nil {
		c.log.Error("Failed to retrieve profile by ID", err)
		return status.UserErrCode_ProfileNotFound, nil, err
	}

	return status.Success_Generic, profile, nil
}
