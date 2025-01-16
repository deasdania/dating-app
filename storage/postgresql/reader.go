package postgresql

import (
	"context"

	"github.com/deasdania/dating-app/storage/models"
	"github.com/google/uuid"
)

//go:generate mockgen -source=writer.go -destination=mock/mock_writer.go
//go:generate gofumpt -s -w mock/mock_writer.go
type IWriterStore interface {
	GetProfiles(ctx context.Context, opts ...models.ProfileFilterOption) ([]*models.Profile, error)
	GetSwipes(ctx context.Context, opts ...models.SwipeFilterOption) ([]*models.Swipe, []*uuid.UUID, error)
	GetUsers(ctx context.Context, opts ...models.UserFilterOption) ([]*models.User, error)
	GetPremiumPackages(ctx context.Context, opts ...models.PremiumPackageFilterOption) ([]*models.PremiumPackage, error)
}
