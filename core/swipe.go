package core

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/deasdania/dating-app/status"
	"github.com/deasdania/dating-app/storage/models"
)

func (c *Core) Swipe(ctx context.Context, req *models.Swipe) (status.DatingStatusCode, error) {
	c.log.Info("Starting swipe core")
	_, err := c.getProfileByID(ctx, req.ProfileID)
	if err != nil {
		c.log.Error("Failed to check user profile", err)
		return status.UserErrCode_ProfileNotFound, err
	}
	return c.processSwipe(ctx, req)
}

// Helper function to process swipe
func (c *Core) processSwipe(ctx context.Context, req *models.Swipe) (status.DatingStatusCode, error) {
	user, err := c.getUserByID(ctx, req.UserID)
	if err != nil {
		c.log.Error("Failed to check user", err)
		return status.SystemErrCode_FailedBrowseData, err
	}

	date := time.Now().Format("2006-01-02")
	_, ids, err := c.storage.GetSwipes(ctx, models.SwipeFilterByCreatedAtDate(date), models.SwipeFilterByUserID(req.UserID), models.SwipeFilterByProfileID(req.ProfileID))
	if len(ids) > 0 {
		c.log.Error("User already swiped on this profile today", err)
		return status.UserErrCode_AlreadySwiped, fmt.Errorf(string(status.UserErrCode_AlreadySwiped))
	}

	swipesCountKey := fmt.Sprintf("swipes_count:%s:%s", req.UserID, date)
	swipesProfilesKey := fmt.Sprintf("swipes_profiles:%s:%s", req.UserID, req.ProfileID)

	swipesCountStr, err := c.getCacheObject(ctx, swipesCountKey)
	if err != nil && err.Error() == "key does not exist" {
		err := c.setCacheObject(ctx, swipesCountKey, "0", 24*time.Hour)
		if err != nil {
			c.log.Error("Failed to initialize swipe count", err)
			return status.SystemErrCode_FailedSwipeTracking, err
		}
		swipesCountStr = "0"
	}

	swipesCount, err := strconv.Atoi(swipesCountStr)
	if err != nil {
		c.log.Error("Failed to convert swipe count", err)
		return status.SystemErrCode_FailedParseSwipe, err
	}

	if !user.IsPremium && swipesCount >= 10 {
		c.log.Error("User has reached daily swipe limit", nil)
		return status.UserErrCode_ReachDailyLimit, fmt.Errorf(string(status.UserErrCode_ReachDailyLimit))
	}

	profileAlreadySwiped, err := c.getCacheObject(ctx, swipesProfilesKey)
	if err == nil && profileAlreadySwiped == req.ProfileID.String() {
		c.log.Error("User already swiped on this profile today", err)
		return status.UserErrCode_AlreadySwiped, fmt.Errorf(string(status.UserErrCode_AlreadySwiped))
	}

	err = c.setCacheObject(ctx, swipesProfilesKey, req.ProfileID, 24*time.Hour)
	if err != nil {
		c.log.Error("Failed to cache swiped profile", err)
		return status.SystemErrCode_FailedSwipeAddingProfile, err
	}

	swipesCount++
	err = c.setCacheObject(ctx, swipesCountKey, fmt.Sprintf("%d", swipesCount), 24*time.Hour)
	if err != nil {
		c.log.Error("Failed to update swipe count in cache", err)
		return status.SystemErrCode_FailedSwipeUpdatingSwipeCount, err
	}

	swipeRecord := models.NewSwipe()
	swipeRecord.UserID = req.UserID
	swipeRecord.ProfileID = req.ProfileID
	swipeRecord.Direction = req.Direction
	_, err = c.storage.CreateSwipe(ctx, swipeRecord)
	if err != nil {
		c.log.Error("Failed to store swipe record", err)
		return status.SystemErrCode_FailedStoreData, err
	}

	return status.Success_Generic, nil
}
