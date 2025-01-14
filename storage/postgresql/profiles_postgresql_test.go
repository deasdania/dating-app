package postgresql

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/deasdania/dating-app/storage/models"
)

func TestStorage_CreateProfile(t *testing.T) {
	envTest(t)
	s, tearDownFn := newTestStorage(t)
	t.Cleanup(tearDownFn)
	ctx := context.Background()

	// Prepare test data
	profileID := uuid.New()
	testData := &models.Profile{
		ID:          &profileID,
		Username:    "testprofile",
		Description: "This is a test profile",
		ImageURL:    "https://example.com/image.jpg",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Create a new profile
	got, err := s.CreateProfile(ctx, testData)
	if err != nil {
		t.Errorf("Storage.CreateProfile() error = %v", err)
		return
	}

	// Validate the returned ID
	uuidProfile, err := uuid.Parse(got.String())
	if err != nil {
		t.Errorf("Storage.CreateProfile() parse id = error %q", err)
	}

	// Fetch the created profile by ID
	filters := []models.ProfileFilterOption{}
	filters = append(filters, models.ProfileFilterByID(&uuidProfile))
	profiles, err := s.GetProfiles(ctx, filters...)
	if err != nil {
		t.Errorf("Storage.GetProfiles() error = %v", err)
	}
	if len(profiles) != 1 {
		t.Errorf("Expected exactly one profile, got %d", len(profiles))
	}
	profile := profiles[0]

	// Validate the fetched profile
	assert.Equal(t, got.String(), profile.ID.String(), "Unexpected ID for profile")

	// Clean and add Username filter
	filters = []models.ProfileFilterOption{}
	filters = append(filters, models.ProfileFilterByUsername(testData.Username))
	profiles, err = s.GetProfiles(ctx, filters...)
	if err != nil {
		t.Errorf("Storage.GetProfiles() error = %v", err)
	}
	assert.Len(t, profiles, 1, "Expected exactly one profile")
	assert.Equal(t, profile.Username, testData.Username, "Expected Username is not matched")
}
