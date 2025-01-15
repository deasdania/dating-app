package core

import (
	"context"
	"errors"
	"time"

	"github.com/deasdania/dating-app/storage/models"
	ps "github.com/deasdania/dating-app/storage/postgresql"
	"github.com/golang-jwt/jwt"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type Core struct {
	log     *logrus.Entry
	storage *ps.Storage
	td      time.Duration
	secret  string
}

const randomListgenerating = 10

// NewCore will create new a Core object representation of ICore interface
func NewCore(log *logrus.Entry, storage *ps.Storage, td time.Duration, secret string) *Core {
	return &Core{
		log:     log,
		storage: storage,
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
		"sub": user.ID.String(), // Store the UUID as a string in the token
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})
	tokenString, err := token.SignedString([]byte(c.secret))
	if err != nil {
		c.log.Error("Error generate token", err)
		return "", err
	}
	return tokenString, nil
}
