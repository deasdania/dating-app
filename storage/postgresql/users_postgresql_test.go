package postgresql

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/deasdania/dating-app/storage/models"
)

func TestStorage_User(t *testing.T) {
	envTest(t)
	s, tearDownFn := newTestStorage(t)
	t.Cleanup(tearDownFn)
	ctx := context.Background()

	// Prepare test data
	userID := uuid.New()
	testData := &models.User{
		ID:        userID,
		Username:  "testuser",
		Password:  "securepassword",
		Email:     "testuser@example.com",
		CreatedAt: time.Now(),
		IsPremium: false,
		Verified:  true,
	}

	// Create a new user
	got, err := s.CreateUser(ctx, testData)
	if err != nil {
		t.Errorf("Storage.CreateUser() error = %v", err)
		return
	}

	// Validate the returned ID
	uuidUser, err := uuid.Parse(got.String())
	if err != nil {
		t.Errorf("Storage.CreateUser() parse id = error %q", err)
	}

	// Fetch the created user by ID
	filters := []models.UserFilterOption{}
	filters = append(filters, models.UserFilterByID(&uuidUser))
	users, err := s.GetUsers(ctx, filters...)
	if err != nil {
		t.Errorf("Storage.GetUsers() error = %v", err)
	}
	if len(users) != 1 {
		t.Errorf("Expected exactly one user, got %d", len(users))
	}
	user := users[0]

	// Validate the fetched user
	assert.Equal(t, got.String(), user.ID.String(), "Unexpected ID for user")

	// Clean and add Username filter
	filters = []models.UserFilterOption{}
	filters = append(filters, models.UserFilterByUsername(testData.Username))
	users, err = s.GetUsers(ctx, filters...)
	if err != nil {
		t.Errorf("Storage.GetUsers() error = %v", err)
	}
	assert.Len(t, users, 1, "Expected exactly one user")
	assert.Equal(t, user.Username, testData.Username, "Expected Username is not matched")
}
