package postgresql

import (
	"context"

	"github.com/deasdania/dating-app/storage/models"
	"github.com/google/uuid"
)

//go:generate mockgen -source=writer.go -destination=mock/mock_reader.go
//go:generate gofumpt -s -w mock/mock_writer.go
type IReaderStore interface {
	CreateProfile(ctx context.Context, profile *models.Profile) (*uuid.UUID, error)
	CreateSwipe(ctx context.Context, swipe *models.Swipe) (*uuid.UUID, error)
	CreateUser(ctx context.Context, user *models.User) (*uuid.UUID, error)
	UpdateProfilePartial(ctx context.Context, profile *models.Profile) error
}
