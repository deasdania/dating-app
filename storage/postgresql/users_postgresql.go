package postgresql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/deasdania/dating-app/storage/models"

	"github.com/google/uuid"
)

const (
	// User Queries
	createUserQuery = `
		INSERT INTO users (
			id,
			username,
			password,
			email,
			created_at,
			is_premium,
			verified,
			daily_swipe_count
		) VALUES (
			:id,
			:username,
			:password,
			:email,
			:created_at,
			:is_premium,
			:verified,
			:daily_swipe_count
		) RETURNING id
	`

	getUserQuery = `
		SELECT 
			id,
			username,
			password,
			email,
			created_at,
			is_premium,
			verified,
			daily_swipe_count
		FROM 
			users
	`

	// User Filter Clauses
	userIDClause          = ` id = :id`
	usernameClause        = ` username = :username`
	emailClause           = ` email = :email`
	isPremiumClause       = ` is_premium = :is_premium`
	verifiedClause        = ` verified = :verified`
	dailySwipeCountClause = ` daily_swipe_count = :daily_swipe_count`
)

func (s *Storage) CreateUser(ctx context.Context, user *models.User) (*uuid.UUID, error) {
	var id uuid.UUID
	stmt, err := s.db.PrepareNamedContext(ctx, createUserQuery)
	if err != nil {
		return nil, fmt.Errorf("preparing named query for createUser: %w", err)
	}
	defer stmt.Close()

	if tx := getTx(ctx); tx != nil {
		stmt, err = tx.Tx.PrepareNamedContext(ctx, createUserQuery)
		if err != nil {
			return nil, fmt.Errorf("error executing query within transaction: %w", err)
		}
	}

	if err := stmt.GetContext(ctx, &id, user); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("failed adding user: %w", err)
		}
		return nil, err
	}
	return &id, nil
}

func (s *Storage) GetUsers(ctx context.Context, opts ...models.UserFilterOption) ([]*models.User, error) {
	filter := &models.UserFilter{}
	for _, opt := range opts {
		opt(filter)
	}

	query, args := buildUserFilter(filter)
	stmt, err := s.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("preparing named query for GetUsers: %w", err)
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Password,
			&user.Email,
			&user.CreatedAt,
			&user.IsPremium,
			&user.Verified,
			&user.DailySwipeCount,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func buildUserFilter(filter *models.UserFilter) (string, map[string]interface{}) {
	query := getUserQuery
	params := make(map[string]interface{})

	if filter.ID != nil {
		query = addQueryString(query, userIDClause)
		params["id"] = filter.ID
	}
	if filter.Username != "" {
		query = addQueryString(query, usernameClause)
		params["username"] = filter.Username
	}
	if filter.Email != "" {
		query = addQueryString(query, emailClause)
		params["email"] = filter.Email
	}
	if filter.IsPremium != nil && *filter.IsPremium {
		query = addQueryString(query, isPremiumClause)
		params["is_premium"] = filter.IsPremium
	}
	if filter.Verified != nil && *filter.Verified {
		query = addQueryString(query, verifiedClause)
		params["verified"] = filter.Verified
	}
	if filter.DailySwipeCount != nil && *filter.DailySwipeCount != 0 {
		query = addQueryString(query, dailySwipeCountClause)
		params["daily_swipe_count"] = filter.DailySwipeCount
	}

	return query, params
}
