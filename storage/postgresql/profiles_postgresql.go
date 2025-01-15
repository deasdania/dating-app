package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/deasdania/dating-app/storage/models"

	"github.com/google/uuid"
)

const (
	// Profile Queries
	createProfileQuery = `
		INSERT INTO profiles (
			id,
			username,
			description,
			image_url,
			created_at,
			updated_at
		) VALUES (
			:id,
			:username,
			:description,
			:image_url,
			:created_at,
			:updated_at
		) RETURNING id
	`

	getProfileQuery = `
		SELECT 
			id,
			username,
			description,
			image_url,
			created_at,
			updated_at
		FROM 
			profiles
	`

	// Profile Filter Clauses
	profileIDClause        = ` id = :id`
	usernameProfileClause  = ` username = :username`
	imageURLClause         = ` image_url = :image_url`
	createdAtProfileClause = ` created_at = :created_at`
	updatedAtProfileClause = ` updated_at = :updated_at`
)

func (s *Storage) CreateProfile(ctx context.Context, profile *models.Profile) (*uuid.UUID, error) {
	var id uuid.UUID
	stmt, err := s.db.PrepareNamedContext(ctx, createProfileQuery)
	if err != nil {
		return nil, fmt.Errorf("preparing named query for createProfile: %w", err)
	}
	defer stmt.Close()

	if tx := getTx(ctx); tx != nil {
		stmt, err = tx.Tx.PrepareNamedContext(ctx, createProfileQuery)
		if err != nil {
			return nil, fmt.Errorf("error executing query within transaction: %w", err)
		}
	}

	if err := stmt.GetContext(ctx, &id, profile); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("failed adding profile: %w", err)
		}
		return nil, err
	}
	return &id, nil
}

func (s *Storage) GetProfiles(ctx context.Context, opts ...models.ProfileFilterOption) ([]*models.Profile, error) {
	filter := &models.ProfileFilter{}
	for _, opt := range opts {
		opt(filter)
	}

	query, args := buildProfileFilter(filter)
	stmt, err := s.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("preparing named query for GetProfiles: %w", err)
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []*models.Profile
	for rows.Next() {
		var profile models.Profile
		err := rows.Scan(
			&profile.ID,
			&profile.Username,
			&profile.Description,
			&profile.ImageURL,
			&profile.CreatedAt,
			&profile.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, &profile)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return profiles, nil
}

func buildProfileFilter(filter *models.ProfileFilter) (string, map[string]interface{}) {
	query := getProfileQuery
	params := make(map[string]interface{})

	if filter.ID != nil {
		query = addQueryString(query, profileIDClause)
		params["id"] = filter.ID
	}
	if filter.Username != "" {
		query = addQueryString(query, usernameProfileClause)
		params["username"] = filter.Username
	}
	if filter.ImageURL != "" {
		query = addQueryString(query, imageURLClause)
		params["image_url"] = filter.ImageURL
	}
	if filter.CreatedAt != nil {
		query = addQueryString(query, createdAtProfileClause)
		params["created_at"] = filter.CreatedAt
	}
	if filter.UpdatedAt != nil {
		query = addQueryString(query, updatedAtProfileClause)
		params["updated_at"] = filter.UpdatedAt
	}

	return query, params
}

func (s *Storage) UpdateProfilePartial(ctx context.Context, profile *models.Profile) error {
	updateParts := []string{}
	params := make(map[string]interface{})

	// Dynamically add fields to update
	if profile.Description != "" {
		updateParts = append(updateParts, "description = :description")
		params["description"] = profile.Description
	}
	if profile.ImageURL != "" {
		updateParts = append(updateParts, "image_url = :image_url")
		params["image_url"] = profile.ImageURL
	}

	// Ensure at least one field to update is provided
	if len(updateParts) == 0 {
		return fmt.Errorf("no fields to update")
	}

	// Combine parts into the final query
	updateQuery := fmt.Sprintf(`
		UPDATE profiles
		SET %s
		WHERE id = :id
	`, strings.Join(updateParts, ", "))

	// Add the profile ID to parameters
	params["id"] = profile.ID

	// Prepare the statement
	stmt, err := s.db.PrepareNamedContext(ctx, updateQuery)
	if err != nil {
		return fmt.Errorf("preparing named query for partial updateProfile: %w", err)
	}
	defer stmt.Close()

	// Execute the query
	_, err = stmt.ExecContext(ctx, params)
	if err != nil {
		return fmt.Errorf("executing partial update query: %w", err)
	}

	return nil
}
