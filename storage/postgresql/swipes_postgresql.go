package postgresql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/deasdania/dating-app/storage/models"

	"github.com/google/uuid"
)

const (
	// Swipe Queries
	createSwipeQuery = `
		INSERT INTO swipes (
			user_id,
			profile_id,
			direction,
			created_at
		) VALUES (
			:user_id,
			:profile_id,
			:direction,
			:created_at
		) RETURNING id
	`

	getSwipeQuery = `
		SELECT 
			id,
			user_id,
			profile_id,
			direction,
			created_at
		FROM 
			swipes
	`

	// Swipe Filter Clauses
	swipesIDClause            = ` id = :id`
	swipesUserIDClause        = ` user_id = :user_id`
	swipesProfileIDClause     = ` profile_id = :profile_id`
	swipesDirectionClause     = ` direction = :direction`
	swipesCreatedAtClause     = ` created_at = :created_at`
	swipesCreatedAtDateClause = ` created_at_date = :created_at_date`
)

func (s *Storage) CreateSwipe(ctx context.Context, swipe *models.Swipe) (*uuid.UUID, error) {
	var id uuid.UUID
	stmt, err := s.db.PrepareNamedContext(ctx, createSwipeQuery)
	if err != nil {
		return nil, fmt.Errorf("preparing named query for createSwipe: %w", err)
	}
	defer stmt.Close()

	if tx := getTx(ctx); tx != nil {
		stmt, err = tx.Tx.PrepareNamedContext(ctx, createSwipeQuery)
		if err != nil {
			return nil, fmt.Errorf("error executing query within transaction: %w", err)
		}
	}

	if err := stmt.GetContext(ctx, &id, swipe); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("failed adding swipe: %w", err)
		}
		return nil, err
	}
	return &id, nil
}

func (s *Storage) GetSwipes(ctx context.Context, opts ...models.SwipeFilterOption) ([]*models.Swipe, []*uuid.UUID, error) {
	filter := &models.SwipeFilter{}
	for _, opt := range opts {
		opt(filter)
	}

	query, args := buildSwipeFilter(filter)
	stmt, err := s.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, nil, fmt.Errorf("preparing named query for GetSwipes: %w", err)
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, args)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var swipes []*models.Swipe
	var swipesProfileIDs []*uuid.UUID
	for rows.Next() {
		var swipe models.Swipe
		err := rows.Scan(
			&swipe.ID,
			&swipe.UserID,
			&swipe.ProfileID,
			&swipe.Direction,
			&swipe.CreatedAt,
		)
		if err != nil {
			return nil, nil, err
		}
		swipes = append(swipes, &swipe)
		swipesProfileIDs = append(swipesProfileIDs, swipe.ProfileID)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return swipes, swipesProfileIDs, nil
}

func buildSwipeFilter(filter *models.SwipeFilter) (string, map[string]interface{}) {
	query := getSwipeQuery
	params := make(map[string]interface{})

	if filter.ID != nil {
		query = addQueryString(query, swipesIDClause)
		params["id"] = filter.ID
	}
	if filter.UserID != nil {
		query = addQueryString(query, swipesUserIDClause)
		params["user_id"] = filter.UserID
	}
	if filter.ProfileID != nil {
		query = addQueryString(query, swipesProfileIDClause)
		params["profile_id"] = filter.ProfileID
	}
	if filter.Direction != "" {
		query = addQueryString(query, swipesDirectionClause)
		params["direction"] = filter.Direction
	}
	if filter.CreatedAt != nil {
		query = addQueryString(query, swipesCreatedAtClause)
		params["created_at"] = filter.CreatedAt
	}
	if filter.CreatedAtDate != "" {
		query = addQueryString(query, swipesCreatedAtDateClause)
		params["created_at_date"] = filter.CreatedAtDate
	}

	return query, params
}
