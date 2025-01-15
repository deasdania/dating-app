package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID              uuid.UUID `json:"id" db:"id"`
	Username        string    `json:"username" db:"username"`
	Password        string    `json:"password" db:"password"`
	Email           string    `json:"email" db:"email"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	IsPremium       bool      `json:"is_premium" db:"is_premium"`
	Verified        bool      `json:"verified" db:"verified"`
	DailySwipeCount int       `json:"daily_swipe_count" db:"daily_swipe_count"`
}

type UserFilter struct {
	ID              *uuid.UUID
	Username        string
	Email           string
	IsPremium       *bool
	Verified        *bool
	DailySwipeCount *int
}

type UserFilterOption func(*UserFilter)

// Define filtering functions for different User attributes

func UserFilterByID(id *uuid.UUID) UserFilterOption {
	return func(f *UserFilter) {
		f.ID = id
	}
}

func UserFilterByUsername(username string) UserFilterOption {
	return func(f *UserFilter) {
		f.Username = username
	}
}

func UserFilterByEmail(email string) UserFilterOption {
	return func(f *UserFilter) {
		f.Email = email
	}
}

func UserFilterByIsPremium(isPremium bool) UserFilterOption {
	return func(f *UserFilter) {
		f.IsPremium = &isPremium
	}
}

func UserFilterByVerified(verified bool) UserFilterOption {
	return func(f *UserFilter) {
		f.Verified = &verified
	}
}

func UserFilterByDailySwipeCount(dailySwipeCount int) UserFilterOption {
	return func(f *UserFilter) {
		f.DailySwipeCount = &dailySwipeCount
	}
}
