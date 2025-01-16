package core

import (
	"context"
	"fmt"

	"github.com/deasdania/dating-app/status"
	"github.com/deasdania/dating-app/storage/models"
	"golang.org/x/crypto/bcrypt"
)

func (c *Core) SignUp(ctx context.Context, user *models.User) (status.DatingStatusCode, error) {
	c.log.Info("starting signup core")

	users, _ := c.storage.GetUsers(ctx, models.UserFilterByEmail(user.Email))
	if users != nil {
		return status.UserErrCode_EmailIsTaken, fmt.Errorf(string(status.UserErrCode_EmailIsTaken))
	}
	users, _ = c.storage.GetUsers(ctx, models.UserFilterByUsername(user.Username))
	if users != nil {
		return status.UserErrCode_UsernameIsTaken, fmt.Errorf(string(status.UserErrCode_UsernameIsTaken))
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.log.Error("Failed to generate hashed password", err)
		return status.SystemErrCode_FailedGenerateHashedPassword, err
	}

	// Save user to database
	user.Password = string(hashedPassword)
	user_id, err := c.storage.CreateUser(ctx, user)
	if err != nil {
		c.log.Error("Failed to create user", err)
		return status.SystemErrCode_FailedCreateUser, err
	}
	c.log.Infof("New user added: %s", user_id)

	return status.Success_Generic, nil
}

func (c *Core) Login(ctx context.Context, input *models.User) (status.DatingStatusCode, string, error) {
	c.log.Info("Starting login core")

	user, err := c.getUserByUsername(ctx, input.Username)
	if err != nil {
		// not found user
		c.log.Error("Failed to get user by username", err)
		return status.UserErrCode_UserNotFound, "", err
	}

	// Check password
	err = c.checkPassword(user.Password, input.Password)
	if err != nil {
		c.log.Error("Failed to compare passwords", err)
		return status.SystemErrCode_FailedCompareHashPassword, "", err
	}

	// Generate JWT token
	tokenString, err := c.generateJWTToken(user.ID.String())
	if err != nil {
		c.log.Error("Failed to generate JWT token", err)
		return status.SystemErrCode_FailedGenerateJWTToken, "", err
	}

	return status.Success_Generic, tokenString, nil
}
