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
	// Premium Package Queries
	createPremiumPackageQuery = `
		INSERT INTO premium_packages (
			id,
			user_id,
			package_type,
			active_until,
			created_at,
			updated_at
		) VALUES (
			:id,
			:user_id,
			:package_type,
			:active_until,
			:created_at,
			:updated_at
		) RETURNING id
	`

	getPremiumPackageQuery = `
		SELECT 
			id,
			user_id,
			package_type,
			active_until,
			created_at,
			updated_at
		FROM 
			premium_packages
	`

	// Premium Package Filter Clauses
	premiumPackageIDClause          = ` id = :id`
	premiumPackageUserIDClause      = ` user_id = :user_id`
	premiumPackageTypeClause        = ` package_type = :package_type`
	premiumPackageActiveUntilClause = ` active_until = :active_until`
	premiumPackageCreatedAtClause   = ` created_at = :created_at`
	premiumPackageUpdatedAtClause   = ` updated_at = :updated_at`
)

func (s *Storage) CreatePremiumPackage(ctx context.Context, premiumPackage *models.PremiumPackage) (*uuid.UUID, error) {
	var id uuid.UUID
	stmt, err := s.db.PrepareNamedContext(ctx, createPremiumPackageQuery)
	if err != nil {
		return nil, fmt.Errorf("preparing named query for createPremiumPackage: %w", err)
	}
	defer stmt.Close()

	if tx := getTx(ctx); tx != nil {
		stmt, err = tx.Tx.PrepareNamedContext(ctx, createPremiumPackageQuery)
		if err != nil {
			return nil, fmt.Errorf("error executing query within transaction: %w", err)
		}
	}

	if err := stmt.GetContext(ctx, &id, premiumPackage); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("failed adding premium package: %w", err)
		}
		return nil, err
	}
	return &id, nil
}

func (s *Storage) GetPremiumPackages(ctx context.Context, opts ...models.PremiumPackageFilterOption) ([]*models.PremiumPackage, error) {
	filter := &models.PremiumPackageFilter{}
	for _, opt := range opts {
		opt(filter)
	}

	query, args := buildPremiumPackageFilter(filter)
	stmt, err := s.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("preparing named query for GetPremiumPackages: %w", err)
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var premiumPackages []*models.PremiumPackage
	for rows.Next() {
		var premiumPackage models.PremiumPackage
		err := rows.Scan(
			&premiumPackage.ID,
			&premiumPackage.UserID,
			&premiumPackage.PackageType,
			&premiumPackage.ActiveUntil,
			&premiumPackage.CreatedAt,
			&premiumPackage.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		premiumPackages = append(premiumPackages, &premiumPackage)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return premiumPackages, nil
}

func buildPremiumPackageFilter(filter *models.PremiumPackageFilter) (string, map[string]interface{}) {
	query := getPremiumPackageQuery
	params := make(map[string]interface{})

	if filter.ID != nil {
		query = addQueryString(query, premiumPackageIDClause)
		params["id"] = filter.ID
	}
	if filter.UserID != nil {
		query = addQueryString(query, premiumPackageUserIDClause)
		params["user_id"] = filter.UserID
	}
	if filter.PackageType != "" {
		query = addQueryString(query, premiumPackageTypeClause)
		params["package_type"] = filter.PackageType
	}
	if filter.ActiveUntil != nil {
		query = addQueryString(query, premiumPackageActiveUntilClause)
		params["active_until"] = filter.ActiveUntil
	}
	if filter.CreatedAt != nil {
		query = addQueryString(query, premiumPackageCreatedAtClause)
		params["created_at"] = filter.CreatedAt
	}
	if filter.UpdatedAt != nil {
		query = addQueryString(query, premiumPackageUpdatedAtClause)
		params["updated_at"] = filter.UpdatedAt
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

func (s *Storage) UpdatePremiumPackagePartial(ctx context.Context, premiumPackage *models.PremiumPackage) error {
	updateParts := []string{}
	params := make(map[string]interface{})

	// Dynamically add fields to update
	if premiumPackage.PackageType != "" {
		updateParts = append(updateParts, "package_type = :package_type")
		params["package_type"] = premiumPackage.PackageType
	}
	if !premiumPackage.ActiveUntil.IsZero() {
		updateParts = append(updateParts, "active_until = :active_until")
		params["active_until"] = premiumPackage.ActiveUntil
	}

	// Ensure at least one field to update is provided
	if len(updateParts) == 0 {
		return fmt.Errorf("no fields to update")
	}

	// Combine parts into the final query
	updateQuery := fmt.Sprintf(`
		UPDATE premium_packages
		SET %s
		WHERE id = :id
	`, strings.Join(updateParts, ", "))

	// Add the premium package ID to parameters
	params["id"] = premiumPackage.ID

	// Prepare the statement
	stmt, err := s.db.PrepareNamedContext(ctx, updateQuery)
	if err != nil {
		return fmt.Errorf("preparing named query for partial updatePremiumPackage: %w", err)
	}
	defer stmt.Close()

	// Execute the query
	_, err = stmt.ExecContext(ctx, params)
	if err != nil {
		return fmt.Errorf("executing partial update query: %w", err)
	}

	return nil
}
