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
			verified
		) VALUES (
			:id,
			:username,
			:password,
			:email,
			:created_at,
			:is_premium,
			:verified
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
			verified
		FROM 
			users
	`

	// User Update Query
	updateUserQuery = `
		UPDATE users
		SET
			username = COALESCE(:username, username),
			password = COALESCE(:password, password),
			email = COALESCE(:email, email),
			is_premium = COALESCE(:is_premium, is_premium),
			verified = COALESCE(:verified, verified),
			updated_at = COALESCE(:updated_at, updated_at)
		WHERE id = :id
		RETURNING id
	`

	// User Filter Clauses
	userIDClause    = ` id = :id`
	usernameClause  = ` username = :username`
	emailClause     = ` email = :email`
	isPremiumClause = ` is_premium = :is_premium`
	verifiedClause  = ` verified = :verified`
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
	return query, params
}

func (s *Storage) UpdateUser(ctx context.Context, user *models.User) error {
	// Prepare the statement for updating user details
	stmt, err := s.db.PrepareNamedContext(ctx, updateUserQuery)
	if err != nil {
		return fmt.Errorf("preparing named query for UpdateUser: %w", err)
	}
	defer stmt.Close()

	// Create a map of the user fields
	params := map[string]interface{}{
		"id":         user.ID,
		"username":   user.Username,
		"password":   user.Password,
		"email":      user.Email,
		"is_premium": user.IsPremium,
		"verified":   user.Verified,
		"updated_at": user.UpdatedAt,
	}

	// Execute the update query
	var updatedID uuid.UUID
	if err := stmt.GetContext(ctx, &updatedID, params); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user with ID %v not found", user.ID)
		}
		return fmt.Errorf("error executing update query: %w", err)
	}

	// Ensure that the returned id matches the ID we intended to update
	if updatedID != user.ID {
		return fmt.Errorf("updated user ID mismatch: expected %v, got %v", user.ID, updatedID)
	}

	return nil
}
