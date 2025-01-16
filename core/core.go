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
	redis "github.com/deasdania/dating-app/storage/redis"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	r "github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type Core struct {
	log     *logrus.Entry
	storage ps.IStore
	cache   *redis.RedisConnection
	td      time.Duration
	secret  string
}

const randomListgenerating = 10

// NewCore will create new a Core object representation of ICore interface
func NewCore(log *logrus.Entry, storage ps.IStore, cache *redis.RedisConnection, td time.Duration, secret string) *Core {
	return &Core{
		log:     log,
		storage: storage,
		cache:   cache,
		td:      td,
		secret:  secret,
	}
}

func (c *Core) Swipe(ctx context.Context, req *models.Swipe) (status.DatingStatusCode, error) {
	c.log.Info("Starting swipe core")
	return c.processSwipe(ctx, req)
}

// Helper function to retrieve user by username
func (c *Core) getUserByID(ctx context.Context, userID *uuid.UUID) (*models.User, error) {
	users, err := c.storage.GetUsers(ctx, models.UserFilterByID(userID))
	if err != nil {
		c.log.Error("Error fetching users from database", err)
		return nil, errors.New("failed to retrieve users")
	}

	if len(users) == 0 {
		c.log.Error("User not found in the database", nil)
		return nil, errors.New("user not found")
	}

	return users[0], nil
}

// Helper function to retrieve user by username
func (c *Core) getUserByUsername(ctx context.Context, username string) (*models.User, error) {
	users, err := c.storage.GetUsers(ctx, models.UserFilterByUsername(username))
	if err != nil {
		c.log.Error("Error fetching users from database", err)
		return nil, errors.New("failed to retrieve users")
	}

	if len(users) == 0 {
		c.log.Error("User not found in the database", nil)
		return nil, errors.New("user not found")
	}

	return users[0], nil
}

// Helper function to compare password hash
func (c *Core) checkPassword(hashedPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		c.log.Error("Password comparison failed", err)
		return fmt.Errorf("Password comparison failed: %v", err)
	}
	return nil
}

// Helper function to generate JWT token
func (c *Core) generateJWTToken(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  userID, // Store the UUID as a string in the token
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(c.secret))
	if err != nil {
		c.log.Error("Error generating JWT token string", err)
		return "", err
	}
	return tokenString, nil
}

// Helper function to retrieve profile by user ID
func (c *Core) getProfileByUserID(ctx context.Context, userID *uuid.UUID) (*models.Profile, error) {
	profiles, err := c.storage.GetProfiles(ctx, models.ProfileFilterByUserID(userID))
	if err != nil {
		c.log.Error("Error fetching profile for user", err)
		return nil, errors.New("failed to retrieve user profile")
	}

	if len(profiles) == 0 {
		c.log.Error("No profile found for user", nil)
		return nil, errors.New("profile not found")
	}

	return profiles[0], nil
}

// Helper function to retrieve profile by profile ID
func (c *Core) getProfileByID(ctx context.Context, profileID *uuid.UUID) (*models.Profile, error) {
	profiles, err := c.storage.GetProfiles(ctx, models.ProfileFilterByID(profileID))
	if err != nil {
		c.log.Error("Error fetching profile by ID", err)
		return nil, errors.New("failed to retrieve profile by ID")
	}

	if len(profiles) == 0 {
		c.log.Error("No profile found with the given ID", nil)
		return nil, errors.New("profile not found")
	}

	return profiles[0], nil
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

// Function to set a cache object with expiration
func (c *Core) setCacheObject(ctx context.Context, key string, value interface{}, duration time.Duration) error {
	// Set the value in Redis
	err := c.cache.Cl.Set(ctx, key, value, duration).Err()
	if err != nil {
		return fmt.Errorf("could not set cache object: %v", err)
	}

	return nil
}

// Function to get a cache object from Redis
func (c *Core) getCacheObject(ctx context.Context, key string) (string, error) {
	// Get the value from Redis
	val, err := c.cache.Cl.Get(ctx, key).Result()
	if err == r.Nil {
		return "", fmt.Errorf("key does not exist")
	} else if err != nil {
		return "", fmt.Errorf("could not get cache object: %v", err)
	}

	return val, nil
}
