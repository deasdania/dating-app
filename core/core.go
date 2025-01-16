package core

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/deasdania/dating-app/status"
	"github.com/deasdania/dating-app/storage/models"
	ps "github.com/deasdania/dating-app/storage/postgresql"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type Core struct {
	log     *logrus.Entry
	storage ps.IStore
	cache   cacheStoreI
	td      time.Duration
	secret  string
}

const randomListgenerating = 10

type cacheStoreI interface {
	SetCacheObject(ctx context.Context, key string, value interface{}, duration time.Duration) error
	GetCacheObject(ctx context.Context, key string) (string, error)
	ReplaceCacheObject(ctx context.Context, key string, value interface{}, duration time.Duration) error
	DeleteCacheObject(ctx context.Context, key string) error
	RenewCacheObjectTimeout(ctx context.Context, key string, duration time.Duration) error
}

// NewCore will create new a Core object representation of ICore interface
func NewCore(log *logrus.Entry, storage ps.IStore, cache cacheStoreI, td time.Duration, secret string) *Core {
	return &Core{
		log:     log,
		storage: storage,
		cache:   cache,
		td:      td,
		secret:  secret,
	}
}

func (c *Core) SignUp(ctx context.Context, user *models.User) error {
	c.log.Info("starting signup core")

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.log.Error("Error generate hashed password", err)
		return err
	}

	// Save user to database
	user.Password = string(hashedPassword)
	user_id, err := c.storage.CreateUser(ctx, user)
	if err != nil {
		c.log.Error("Failed to create user")
		return err
	}
	c.log.Infof("new user added %s", user_id)

	return nil
}

func (c *Core) Login(ctx context.Context, input *models.User) (string, error) {
	c.log.Info("starting login core")

	users, err := c.storage.GetUsers(ctx, models.UserFilterByUsername(input.Username))
	var user *models.User
	if users != nil && len(users) > 0 {
		user = users[0] // username is unique, so it should be only one for each
	} else {
		c.log.Error("User not found", err)
		return "", errors.New("not found")
	}

	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		c.log.Error("Error compare passwords", err)
		return "", err
	}

	// Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  user.ID.String(), // Store the UUID as a string in the token
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})
	tokenString, err := token.SignedString([]byte(c.secret))
	if err != nil {
		c.log.Error("Error generate token", err)
		return "", err
	}
	return tokenString, nil
}

func (c *Core) GetProfile(ctx context.Context, userID *uuid.UUID) (*models.Profile, error) {
	c.log.Info("starting get profile core")

	users, err := c.storage.GetUsers(ctx, models.UserFilterByID(userID))
	var user *models.User
	if users != nil && len(users) > 0 {
		user = users[0] // username is unique, so it should be only one for each
	} else {
		c.log.Error("User not found", err)
		return nil, errors.New("not found")
	}
	profiles, err := c.storage.GetProfiles(ctx, models.ProfileFilterByUsername(user.Username))
	var profile *models.Profile
	if profiles != nil && len(profiles) > 0 {
		profile = profiles[0] // username is unique, so it should be only one for each
	} else {
		c.log.Error("User not found", err)
		return nil, errors.New("not found")
	}

	return profile, nil
}

func (c *Core) SetProfile(ctx context.Context, userID *uuid.UUID, req *models.Profile) (bool, error) {
	c.log.Info("starting get profile core")

	isNew := false
	users, err := c.storage.GetUsers(ctx, models.UserFilterByID(userID))
	var user *models.User
	if users != nil && len(users) > 0 {
		user = users[0] // username is unique, so it should be only one for each
	} else {
		c.log.Error("User not found", err)
		return isNew, errors.New("not found")
	}
	profiles, err := c.storage.GetProfiles(ctx, models.ProfileFilterByUsername(user.Username))
	var profile *models.Profile
	if profiles != nil && len(profiles) > 0 {
		profile = profiles[0] // username is unique, so it should be only one for each
		if req.ImageURL == "" {
			req.ImageURL = profile.ImageURL
		}
		if err := c.storage.UpdateProfilePartial(ctx, &models.Profile{
			ID:          profile.ID,
			Description: req.Description,
			ImageURL:    req.ImageURL,
		}); err != nil {
			c.log.Error("Failed update profile", err)
			return isNew, errors.New("failed update profile")
		}
	} else {
		req.Username = user.Username
		isNew = true
		if _, err := c.storage.CreateProfile(ctx, req); err != nil {
			c.log.Error("Failed create profile", err)
			return isNew, errors.New("failed create profile")
		}
	}

	return isNew, nil
}

func (c *Core) GetPeopleProfiles(ctx context.Context, page, limit uint) ([]*models.Profile, error) {
	c.log.Info("starting get people profiles core")

	profiles, err := c.storage.GetProfiles(ctx, models.ProfileFilterByPage(page), models.ProfileFilterByLimit(limit))
	if err != nil {
		c.log.Error("Failed get profiles", err)
		return nil, errors.New("failed get profiles")
	}
	return profiles, nil
}

func (c *Core) GetPeopleProfileByID(ctx context.Context, profileID *uuid.UUID) (*models.Profile, error) {
	c.log.Info("starting get profile by id core")

	profiles, err := c.storage.GetProfiles(ctx, models.ProfileFilterByID(profileID))
	var profile *models.Profile
	if profiles != nil && len(profiles) > 0 {
		profile = profiles[0] // username is unique, so it should be only one for each
	} else {
		c.log.Error("User not found", err)
		return nil, errors.New("not found")
	}

	return profile, nil
}
func (c *Core) Swipe(ctx context.Context, req *models.Swipe) (status.DatingStatusCode, error) {
	// Get the current date in the format YYYY-MM-DD
	date := time.Now().Format("2006-01-02")

	// Redis keys for tracking swipes for the current day
	swipesCountKey := fmt.Sprintf("swipes_count:%s:%s", req.UserID, date)
	swipesProfilesKey := fmt.Sprintf("swipes_profiles:%s:%s", req.UserID, date)

	// Check if the swipe count key exists, if not, create it with a default value of 0
	swipesCountStr, err := c.cache.GetCacheObject(ctx, swipesCountKey)
	if err != nil && err.Error() == "key does not exist" {
		// If the key doesn't exist, initialize the swipe count to 0
		err := c.cache.SetCacheObject(ctx, swipesCountKey, "0", 24*time.Hour)
		if err != nil {
			return status.SystemErrCode_FailedSwipeTracking, err
		}
		swipesCountStr = "0" // No swipes so far today, so it's 0
	}

	// Convert swipe count to integer
	swipesCount, err := strconv.Atoi(swipesCountStr)
	if err != nil {
		return status.SystemErrCode_FailedParseSwipe, err
	}

	// Check if user has exceeded the swipe limit (10 per day)
	if swipesCount >= 10 {
		return status.UserErrCode_ReachDailyLimit, nil
	}

	// Check if the profile has already been swiped on today
	profileAlreadySwiped, err := c.cache.GetCacheObject(ctx, swipesProfilesKey)
	if err == nil && profileAlreadySwiped == req.ProfileID.String() {
		// If the profile is already swiped today, return the appropriate message
		return status.UserErrCode_AlreadySwiped, nil
	} else if err != nil && err.Error() == "key does not exist" {
		// If the key doesn't exist, it means no profile has been swiped yet for today, so proceed.
	}

	// Add the profile ID to the set of profiles swiped by the user today
	err = c.cache.SetCacheObject(ctx, swipesProfilesKey, req.ProfileID, 24*time.Hour)
	if err != nil {
		return status.SystemErrCode_FailedSwipeAddingProfile, err
	}

	// Increment the swipe count
	swipesCount++
	err = c.cache.SetCacheObject(ctx, swipesCountKey, fmt.Sprintf("%d", swipesCount), 24*time.Hour)
	if err != nil {
		return status.SystemErrCode_FailedSwipeUpdatingSwipeCount, err
	}

	// Optionally, log the swipe direction (left or right) for analytics
	fmt.Printf("User %s swiped %s on profile %s\n", req.UserID, req.Direction, req.ProfileID)

	// Set the expiration time for the Redis keys (this will automatically expire at midnight)
	expiration := getNextMidnight() // Calculate the time until midnight
	err = c.cache.RenewCacheObjectTimeout(ctx, swipesCountKey, expiration.Sub(time.Now()))
	if err != nil {
		return status.SystemErrCode_FailedSwipeSettingExpire, err
	}

	err = c.cache.RenewCacheObjectTimeout(ctx, swipesProfilesKey, expiration.Sub(time.Now()))
	if err != nil {
		return status.SystemErrCode_FailedSwipeSettingExpire, err
	}

	// Store the swipe in the database
	_, err = c.storage.CreateSwipe(ctx, &models.Swipe{
		UserID:    req.UserID,
		ProfileID: req.ProfileID,
		Direction: req.Direction,
	})
	if err != nil {
		return status.SystemErrCode_FailedStoreData, err
	}

	return status.Success_Generic, nil
}

// Helper function to get the next midnight timestamp
func getNextMidnight() time.Time {
	now := time.Now()
	loc := now.Location()
	// Calculate the midnight of the next day
	nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, loc)
	return nextMidnight
}
