package models

import (
	"time"

	"github.com/google/uuid"
)

type Swipe struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	UserID    *uuid.UUID `json:"user_id" db:"user_id"`
	ProfileID *uuid.UUID `json:"profile_id" db:"profile_id"`
	Direction string     `json:"direction" db:"direction"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

func NewSwipe() *Swipe {
	return &Swipe{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
	}
}

type SwipeFilter struct {
	ID            *uuid.UUID
	UserID        *uuid.UUID
	ProfileID     *uuid.UUID
	Direction     string
	CreatedAt     *time.Time
	CreatedAtDate string
}

type SwipeFilterOption func(*SwipeFilter)

// Define filtering functions for different Swipe attributes

func SwipeFilterByID(id *uuid.UUID) SwipeFilterOption {
	return func(f *SwipeFilter) {
		f.ID = id
	}
}

func SwipeFilterByUserID(userID *uuid.UUID) SwipeFilterOption {
	return func(f *SwipeFilter) {
		f.UserID = userID
	}
}

func SwipeFilterByProfileID(profileID *uuid.UUID) SwipeFilterOption {
	return func(f *SwipeFilter) {
		f.ProfileID = profileID
	}
}

func SwipeFilterByDirection(direction string) SwipeFilterOption {
	return func(f *SwipeFilter) {
		f.Direction = direction
	}
}

func SwipeFilterByCreatedAt(createdAt *time.Time) SwipeFilterOption {
	return func(f *SwipeFilter) {
		f.CreatedAt = createdAt
	}
}

func SwipeFilterByCreatedAtDate(date string) SwipeFilterOption {
	return func(f *SwipeFilter) {
		f.CreatedAtDate = date
	}
}
