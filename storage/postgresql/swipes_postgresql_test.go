package postgresql

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/deasdania/dating-app/storage/models"
)

func TestStorage_Swipe(t *testing.T) {
	envTest(t)
	s, tearDownFn := newTestStorage(t)
	t.Cleanup(tearDownFn)
	ctx := context.Background()

	// Prepare test data user
	testDataUser := &models.User{
		Username:  "testuser",
		Password:  "securepassword",
		Email:     "testuser@example.com",
		CreatedAt: time.Now(),
		IsPremium: false,
		Verified:  true,
	}

	// Create a new user
	userID, err := s.CreateUser(ctx, testDataUser)
	if err != nil {
		t.Errorf("Storage.CreateUser() error = %v", err)
		return
	}

	// Prepare test data profile
	testDataProfile := &models.Profile{
		UserID:      *userID,
		Username:    "testprofile",
		Description: "This is a test profile",
		ImageURL:    "https://example.com/image.jpg",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	// Create a new profile
	profileID, err := s.CreateProfile(ctx, testDataProfile)
	if err != nil {
		t.Errorf("Storage.CreateProfile() error = %v", err)
		return
	}

	swipeID := uuid.New()
	testDataSwipe := &models.Swipe{
		ID:        swipeID,
		UserID:    userID,
		ProfileID: profileID,
		Direction: "like",
		CreatedAt: time.Now(),
	}

	// Create a new swipe
	got, err := s.CreateSwipe(ctx, testDataSwipe)
	if err != nil {
		t.Errorf("Storage.CreateSwipe() error = %v", err)
		return
	}

	// Validate the returned ID
	uuidSwipe, err := uuid.Parse(got.String())
	if err != nil {
		t.Errorf("Storage.CreateSwipe() parse id = error %q", err)
	}

	// Fetch the created swipe by ID
	filters := []models.SwipeFilterOption{}
	filters = append(filters, models.SwipeFilterByID(&uuidSwipe))
	swipes, _, err := s.GetSwipes(ctx, filters...)
	if err != nil {
		t.Errorf("Storage.GetSwipes() error = %v", err)
	}
	if len(swipes) != 1 {
		t.Errorf("Expected exactly one swipe, got %d", len(swipes))
	}
	swipe := swipes[0]

	// Validate the fetched swipe
	assert.Equal(t, got.String(), swipe.ID.String(), "Unexpected ID for swipe")

	// Clean and add Direction filter
	filters = []models.SwipeFilterOption{}
	filters = append(filters, models.SwipeFilterByDirection(testDataSwipe.Direction))
	swipes, _, err = s.GetSwipes(ctx, filters...)
	if err != nil {
		t.Errorf("Storage.GetSwipes() error = %v", err)
	}
	assert.Len(t, swipes, 1, "Expected exactly one swipe")
	assert.Equal(t, swipe.Direction, testDataSwipe.Direction, "Expected Direction is not matched")
}
