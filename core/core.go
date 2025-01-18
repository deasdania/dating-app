package core

import (
	"context"
	"errors"
	"fmt"
	"time"

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
	timeout int
}

const randomListgenerating = 10

// NewCore will create new a Core object representation of ICore interface
func NewCore(log *logrus.Entry, storage ps.IStore, cache *redis.RedisConnection, td time.Duration, secret string, timeout int) *Core {
	return &Core{
		log:     log,
		storage: storage,
		cache:   cache,
		td:      td,
		secret:  secret,
		timeout: timeout,
	}
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
		"exp": time.Now().Add(time.Minute * time.Duration(c.timeout)).Unix(),
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
