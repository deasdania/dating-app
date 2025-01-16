package postgresql

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/deasdania/dating-app/storage/models"
)

func TestStorage_PremiumPackage(t *testing.T) {
	envTest(t)
	s, tearDownFn := newTestStorage(t)
	t.Cleanup(tearDownFn)
	ctx := context.Background()

	// Step 1: Create a valid user
	userID := uuid.New()
	user := &models.User{
		ID:        userID,
		Username:  "testuser",
		Email:     "testuser@example.com",
		Password:  "password123", // Assuming the password is required
		CreatedAt: time.Now(),
	}

	// Insert the user into the database (you need to have a CreateUser method in your storage)
	_, err := s.CreateUser(ctx, user) // Replace this with actual method to insert a user
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Step 2: Prepare test data for PremiumPackage
	premiumPackageID := uuid.New()
	testData := &models.PremiumPackage{
		ID:          premiumPackageID,
		UserID:      &userID,                             // Use the valid userID from the user created above
		PackageType: models.VerifiedLabel,                // Example package type
		ActiveUntil: time.Now().Add(30 * 24 * time.Hour), // Active for 30 days
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Step 3: Create a new premium package
	got, err := s.CreatePremiumPackage(ctx, testData)
	if err != nil {
		t.Errorf("Storage.CreatePremiumPackage() error = %v", err)
		return
	}

	// Validate the returned ID
	uuidPremiumPackage, err := uuid.Parse(got.String())
	if err != nil {
		t.Errorf("Storage.CreatePremiumPackage() parse id = error %q", err)
	}

	// Fetch the created premium package by ID
	filters := []models.PremiumPackageFilterOption{}
	filters = append(filters, models.PremiumPackageFilterByID(&uuidPremiumPackage))
	premiumPackages, err := s.GetPremiumPackages(ctx, filters...)
	if err != nil {
		t.Errorf("Storage.GetPremiumPackages() error = %v", err)
	}
	if len(premiumPackages) != 1 {
		t.Errorf("Expected exactly one premium package, got %d", len(premiumPackages))
	}
	premiumPackage := premiumPackages[0]

	// Validate the fetched premium package
	assert.Equal(t, got.String(), premiumPackage.ID.String(), "Unexpected ID for premium package")

	// Clean and add UserID filter
	filters = []models.PremiumPackageFilterOption{}
	filters = append(filters, models.PremiumPackageFilterByUserID(testData.UserID))
	premiumPackages, err = s.GetPremiumPackages(ctx, filters...)
	if err != nil {
		t.Errorf("Storage.GetPremiumPackages() error = %v", err)
	}
	assert.Len(t, premiumPackages, 1, "Expected exactly one premium package")
	assert.Equal(t, premiumPackage.UserID.String(), testData.UserID.String(), "Expected UserID is not matched")

	// Update the ActiveUntil field of the premium package
	newActiveUntil := time.Now().Add(60 * 24 * time.Hour) // New expiration date
	premiumPackage.ActiveUntil = newActiveUntil
	err = s.UpdatePremiumPackagePartial(ctx, premiumPackage)
	if err != nil {
		t.Errorf("Storage.UpdatePremiumPackagePartial() error = %v", err)
	}

	// Fetch the updated premium package by ID
	premiumPackages, err = s.GetPremiumPackages(ctx, filters...)
	if err != nil {
		t.Errorf("Storage.GetPremiumPackages() error = %v", err)
	}
	if len(premiumPackages) != 1 {
		t.Errorf("Expected exactly one premium package, got %d", len(premiumPackages))
	}
	premiumPackage = premiumPackages[0]
	assert.Equal(t, premiumPackage.ActiveUntil.Format("2006-01-02"), newActiveUntil.Format("2006-01-02"), "Expected ActiveUntil is not matched")
}
