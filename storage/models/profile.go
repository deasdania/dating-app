package models

import (
	"time"

	"github.com/google/uuid"
)

type Profile struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	Username    string    `json:"username" db:"username"`
	Description string    `json:"description" db:"description"`
	ImageURL    string    `json:"image_url" db:"image_url"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`

	// to show premium and verified flag from users table
	IsPremium bool `json:"is_premium" db:"is_premium"`
	Verified  bool `json:"verified" db:"verified"`
}

type ProfileFilter struct {
	ID                *uuid.UUID
	UserID            *uuid.UUID
	Username          string
	Description       string
	ImageURL          string
	Page              uint
	Limit             uint
	ExcludeProfileIDs []*uuid.UUID
	CreatedAt         *time.Time
	UpdatedAt         *time.Time
}

type ProfileFilterOption func(*ProfileFilter)

// Define filtering functions for different Profile attributes

func ProfileFilterByID(id *uuid.UUID) ProfileFilterOption {
	return func(f *ProfileFilter) {
		f.ID = id
	}
}

func ProfileFilterByUserID(id *uuid.UUID) ProfileFilterOption {
	return func(f *ProfileFilter) {
		f.UserID = id
	}
}

func ProfileFilterByUsername(username string) ProfileFilterOption {
	return func(f *ProfileFilter) {
		f.Username = username
	}
}

func ProfileFilterByDescription(description string) ProfileFilterOption {
	return func(f *ProfileFilter) {
		f.Description = description
	}
}

func ProfileFilterByImageURL(imageURL string) ProfileFilterOption {
	return func(f *ProfileFilter) {
		f.ImageURL = imageURL
	}
}

func ProfileFilterByCreatedAt(createdAt *time.Time) ProfileFilterOption {
	return func(f *ProfileFilter) {
		f.CreatedAt = createdAt
	}
}

func ProfileFilterByUpdatedAt(updatedAt *time.Time) ProfileFilterOption {
	return func(f *ProfileFilter) {
		f.UpdatedAt = updatedAt
	}
}

func ProfileFilterByPage(input uint) ProfileFilterOption {
	return func(f *ProfileFilter) {
		f.Page = input
	}
}

func ProfileFilterByLimit(input uint) ProfileFilterOption {
	return func(f *ProfileFilter) {
		f.Limit = input
	}
}

func ProfileFilterByExcludeProfileIDs(input []*uuid.UUID) ProfileFilterOption {
	return func(f *ProfileFilter) {
		f.ExcludeProfileIDs = input
	}
}
