package core

import (
	"context"

	"github.com/deasdania/dating-app/status"
	"github.com/deasdania/dating-app/storage/models"
	"github.com/google/uuid"
)

func (c *Core) SetPremium(ctx context.Context, userID *uuid.UUID, packageType models.PackageType) (status.DatingStatusCode, error) {
	c.log.Info("Starting set premium core")

	user, err := c.getUserByID(ctx, userID)
	if err != nil {
		c.log.Error("Failed to check user", err)
		return status.SystemErrCode_FailedBrowseData, err
	}

	if packageType == models.VerifiedLabel && user.Verified {
		// check the premium package expiration
	}
	if packageType == models.RemoveQuota && user.IsPremium {
		// check the premium package expiration
	}

	switch packageType {
	case models.VerifiedLabel:
		user.Verified = true
	case models.RemoveQuota:
		user.IsPremium = true
	}
	req := models.NewPremiumPackage()
	req.UserID = userID
	req.PackageType = packageType

	_, err = c.storage.CreatePremiumPackage(ctx, req)
	if err != nil {
		c.log.Error("Failed to store premium package", err)
		return status.SystemErrCode_FailedStoreData, err
	}
	err = c.storage.UpdateUser(ctx, user)
	if err != nil {
		c.log.Error("Failed to store premium package", err)
		return status.SystemErrCode_FailedStoreData, err
	}
	return status.Success_Generic, nil
}
