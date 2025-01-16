package models

import (
	"time"

	"github.com/google/uuid"
)

// Enum type for package type
type PackageType string

const (
	RemoveQuota   PackageType = "remove_quota"
	VerifiedLabel PackageType = "verified_label"
)

type PremiumPackage struct {
	ID          uuid.UUID   `json:"id" db:"id"`                     // Unique ID for each premium package
	UserID      uuid.UUID   `json:"user_id" db:"user_id"`           // User ID as a foreign key
	PackageType PackageType `json:"package_type" db:"package_type"` // Type of the premium package
	ActiveUntil time.Time   `json:"active_until" db:"active_until"` // Date until the package is active
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`     // Timestamp for record creation
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`     // Timestamp for record updates
}

// Filter struct for querying `premium_packages`
type PremiumPackageFilter struct {
	ID          *uuid.UUID
	UserID      *uuid.UUID
	PackageType PackageType
	ActiveUntil *time.Time
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
	Page        uint
	Limit       uint
}

// Define a function type for filtering options
type PremiumPackageFilterOption func(*PremiumPackageFilter)

// Filtering functions for `premium_packages` attributes

func PremiumPackageFilterByID(id *uuid.UUID) PremiumPackageFilterOption {
	return func(f *PremiumPackageFilter) {
		f.ID = id
	}
}

func PremiumPackageFilterByUserID(userID *uuid.UUID) PremiumPackageFilterOption {
	return func(f *PremiumPackageFilter) {
		f.UserID = userID
	}
}

func PremiumPackageFilterByPackageType(packageType PackageType) PremiumPackageFilterOption {
	return func(f *PremiumPackageFilter) {
		f.PackageType = packageType
	}
}

func PremiumPackageFilterByActiveUntil(activeUntil *time.Time) PremiumPackageFilterOption {
	return func(f *PremiumPackageFilter) {
		f.ActiveUntil = activeUntil
	}
}

func PremiumPackageFilterByCreatedAt(createdAt *time.Time) PremiumPackageFilterOption {
	return func(f *PremiumPackageFilter) {
		f.CreatedAt = createdAt
	}
}

func PremiumPackageFilterByUpdatedAt(updatedAt *time.Time) PremiumPackageFilterOption {
	return func(f *PremiumPackageFilter) {
		f.UpdatedAt = updatedAt
	}
}

func PremiumPackageFilterByPage(page uint) PremiumPackageFilterOption {
	return func(f *PremiumPackageFilter) {
		f.Page = page
	}
}

func PremiumPackageFilterByLimit(limit uint) PremiumPackageFilterOption {
	return func(f *PremiumPackageFilter) {
		f.Limit = limit
	}
}
