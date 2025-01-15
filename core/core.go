package core

import (
	"context"
	"time"

	"github.com/deasdania/dating-app/storage/models"
	ps "github.com/deasdania/dating-app/storage/postgresql"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type Core struct {
	log     *logrus.Entry
	storage *ps.Storage
	td      time.Duration
}

const randomListgenerating = 10

// NewCore will create new a Core object representation of ICore interface
func NewCore(log *logrus.Entry, storage *ps.Storage, td time.Duration) *Core {
	return &Core{
		log:     log,
		storage: storage,
		td:      td,
	}
}

func (c *Core) SignUp(ctx context.Context, user *models.User) error {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.log.Error("Error generate hashed password", err)
		return err
	}

	user.Password = string(hashedPassword)
	user_id, err := c.storage.CreateUser(ctx, user)
	// Save user to database
	if err != nil {
		c.log.Error("Failed to create user")
		return err
	}
	c.log.Infof("new user added %s", user_id)

	return nil
}
