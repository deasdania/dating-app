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
			user_id,
			username,
			description,
			image_url,
			created_at,
			updated_at
		) VALUES (
			:id,
			:user_id,
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
			user_id,
			username,
			COALESCE(description, '') AS description,
			COALESCE(image_url, '') AS image_url,created_at,
			updated_at
		FROM 
			profiles
	`

	// Profile Filter Clauses
	profileIDClause        = ` id = :id`
	profileUserIDClause    = ` user_id = :user_id`
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
			&profile.UserID,
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
	if filter.UserID != nil {
		query = addQueryString(query, profileUserIDClause)
		params["user_id"] = filter.UserID
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

	// Exclude profile IDs if they're provided (handling slice of pointers)
	if len(filter.ExcludeProfileIDs) > 0 {
		var ids string
		for i, id := range filter.ExcludeProfileIDs {
			ids += fmt.Sprintf("'%s'", id)
			if i != len(filter.ExcludeProfileIDs)-1 {
				ids += ","
			}
		}
		query = addQueryString(query, fmt.Sprintf("id NOT IN (%s)", ids))
	}

	// Handle pagination: Page and Limit
	page := filter.Page
	limit := filter.Limit

	// Override with filter values, if provided
	if filter.Page == 0 {
		page = 1
	}

	if filter.Limit == 0 {
		limit = 10
	}

	// Calculate the offset based on the page and limit
	offset := (page - 1) * limit // Offset for the query

	// Add pagination (LIMIT and OFFSET) to the query
	query += fmt.Sprintf(" LIMIT :limit OFFSET :offset")
	params["limit"] = limit
	params["offset"] = offset

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

// only for not premium member
func (s *Storage) GetAvailableProfilesForSwiping(ctx context.Context, userID *uuid.UUID, targetDate string) ([]*models.Profile, error) {
	// First, count how many swipes the user has already made today (like + pass)
	countQuery := `
		SELECT COUNT(*)
		FROM swipes
		WHERE user_id = :user_id
		  AND created_at::date = :target_date::date;
	`

	var currentSwipeCount int
	params := map[string]interface{}{
		"user_id":     userID,
		"target_date": targetDate, // Use string formatted as 'YYYY-MM-DD'
	}

	err := s.db.GetContext(ctx, &currentSwipeCount, countQuery, params)
	if err != nil {
		return nil, fmt.Errorf("counting swipes: %w", err)
	}

	// Limit the number of remaining swipes (user can only swipe 10 profiles per day)
	remainingSwipes := 10 - currentSwipeCount
	if remainingSwipes <= 0 {
		// If no swipes are left, return an empty slice
		return []*models.Profile{}, nil
	}

	// Now get the profiles excluding already swiped ones
	profilesQuery := `
		SELECT *
		FROM profiles
		WHERE id NOT IN (
			SELECT profile_id
			FROM swipes
			WHERE user_id = :user_id
			  AND created_at::date = :target_date::date
		)
		  AND id != :user_id
		LIMIT :remaining_swipes;
	`

	// Prepare the parameters for the second query
	params["remaining_swipes"] = remainingSwipes

	// Query the database
	stmt, err := s.db.PrepareNamedContext(ctx, profilesQuery)
	if err != nil {
		return nil, fmt.Errorf("preparing query: %w", err)
	}
	defer stmt.Close()

	// Get profiles from the database
	rows, err := stmt.QueryContext(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("querying profiles: %w", err)
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
			return nil, fmt.Errorf("scanning row: %w", err)
		}
		profiles = append(profiles, &profile)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return profiles, nil
}
