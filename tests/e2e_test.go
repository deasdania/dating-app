package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/deasdania/dating-app/storage/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const (
	USER_ID_TEST = "d5dba9f0-2daf-47d3-9763-f98b1ea25376"
	baseURL      = "http://localhost:8080/v1"
)

func toUID(id string) uuid.UUID {
	i, _ := uuid.Parse(id)
	return i
}

func getAvailableProfileIDByUserID(t *testing.T, userID string) *uuid.UUID {
	connStr := "postgres://postgres:secret@localhost:5432/dating_app?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()
	var profileID *uuid.UUID
	err = db.QueryRow(`SELECT p.id as profile_id
	FROM profiles p
	WHERE p.id NOT IN (
		SELECT s.profile_id
		FROM swipes s
		WHERE s.user_id = $1 
		AND DATE(s.created_at) = CURRENT_DATE
	) order by id
	LIMIT 1;
`, userID).Scan(&profileID)
	if err != nil {
		if err == sql.ErrNoRows {
			t.Error("No profile found for user ID:", userID)
		} else {
			t.Error(err)
		}
	}
	return profileID
}

// Helper function to send POST requests
func sendPostRequest(url string, payload interface{}, token string) (*http.Response, error) {
	// Convert the payload into JSON
	fmt.Println(url)
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	// Set the Content-Type header to application/json
	req.Header.Set("Content-Type", "application/json")

	// Add Authorization header if token is not empty
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	// Create a client and send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// Helper function to send GET requests
func sendGetRequest(url string) (*http.Response, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func TestSignUp(t *testing.T) {
	// Seed the random number generator to get different results each time
	rand.Seed(time.Now().UnixNano())

	// Generate a random number between 1 and 100
	randomNumber := rand.Intn(100) + 1
	username := fmt.Sprintf("testuserseed%dTest", randomNumber)
	fmt.Println(username)
	// Prepare test data for user registration
	user := &models.User{
		Username: username,
		Email:    fmt.Sprintf("%s@example.com", username),
		Password: "Password123",
	}

	// Send the POST request
	resp, err := sendPostRequest(baseURL+"/signup", user, "")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestLogin(t *testing.T) {
	// Prepare test data for login
	loginData := map[string]string{
		"username": "testuser",
		"password": "Password123",
	}

	// Send POST request for login
	resp, err := sendPostRequest(baseURL+"/login", loginData, "")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify that a token is returned
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	token, exists := result["token"]
	assert.True(t, exists, "Token should be returned")
	assert.NotNil(t, token)
}

func TestLoginSwipe(t *testing.T) {
	// Assuming the user has already logged in and received a token
	loginData := map[string]string{
		"username": "user1",
		"password": "password1",
	}
	// Send POST request for login
	resp, err := sendPostRequest(baseURL+"/login", loginData, "")
	assert.NoError(t, err)
	var resultLog map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&resultLog)
	var token string
	token, exists := resultLog["token"].(string)
	assert.True(t, exists, "Token should be returned")
	assert.NotNil(t, token)

	profileID := getAvailableProfileIDByUserID(t, USER_ID_TEST)
	assert.NotNil(t, profileID)
	fmt.Println(profileID)

	// Prepare data for swipe action
	swipe := &models.Swipe{
		ProfileID: profileID,
		Direction: "right", // Example: "left" or "right"
	}

	// Send POST request to perform swipe
	resp, err = sendPostRequest(baseURL+"/swipe", swipe, token)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var resultLogN map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&resultLogN)
}
