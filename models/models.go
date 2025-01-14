package models

import (
	"time"

	"github.com/deasdania/dating-app/status"
	uuid "github.com/google/uuid"
)

type User struct {
	ID              *uuid.UUID `json:"id" gorm:"primary_key;type:uuid;default:uuid_generate_v4()"`
	Username        string     `json:"username" gorm:"unique"`
	Password        string     `json:"password"`
	Email           string     `json:"email" gorm:"unique"`
	CreatedAt       time.Time  `json:"created_at"`
	IsPremium       bool       `json:"is_premium"`
	Verified        bool       `json:"verified"`
	DailySwipeCount int        `json:"daily_swipe_count"`
}

type Profile struct {
	ID          *uuid.UUID `json:"id" gorm:"primary_key;type:uuid;default:uuid_generate_v4()"`
	Username    string     `json:"username"`
	Description string     `json:"description"`
	ImageURL    string     `json:"image_url"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type Swipe struct {
	ID        *uuid.UUID `json:"id" gorm:"primary_key;type:uuid;default:uuid_generate_v4()"`
	UserID    *uuid.UUID `json:"user_id"`
	ProfileID *uuid.UUID `json:"profile_id"`
	Direction string     `json:"direction"` // "like" or "pass"
	CreatedAt time.Time  `json:"created_at"`
}

type ResponseBase struct {
	Status  int64                 `json:"status"`
	Details status.StatusResponse `json:"details"`
	Data    interface{}           `json:"data"`
}
