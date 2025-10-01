package adapters

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/quantfidential/trading-ecosystem/custodian-data-adapter-go/pkg/interfaces"
	"github.com/quantfidential/trading-ecosystem/custodian-data-adapter-go/pkg/models"
	"github.com/sirupsen/logrus"
)

type PostgresPositionRepository struct {
	db     *sql.DB
	logger *logrus.Logger
}

func NewPostgresPositionRepository(db *sql.DB, logger *logrus.Logger) interfaces.PositionRepository {
	return &PostgresPositionRepository{
		db:     db,
		logger: logger,
	}
}

func (r *PostgresPositionRepository) Create(ctx context.Context, position *models.Position) error {
	query := `
		INSERT INTO custodian.positions (
			position_id, account_id, symbol, quantity, available_quantity, locked_quantity,
			average_cost, market_value, currency, last_updated, created_at, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := r.db.ExecContext(ctx, query,
		position.PositionID, position.AccountID, position.Symbol, position.Quantity,
		position.AvailableQuantity, position.LockedQuantity, position.AverageCost,
		position.MarketValue, position.Currency, position.LastUpdated, position.CreatedAt,
		position.Metadata,
	)

	if err != nil {
		r.logger.WithError(err).Error("Failed to create position")
		return fmt.Errorf("failed to create position: %w", err)
	}

	return nil
}

func (r *PostgresPositionRepository) GetByID(ctx context.Context, positionID string) (*models.Position, error) {
	query := `
		SELECT position_id, account_id, symbol, quantity, available_quantity, locked_quantity,
			   average_cost, market_value, currency, last_updated, created_at, metadata
		FROM custodian.positions
		WHERE position_id = $1
	`

	position := &models.Position{}
	err := r.db.QueryRowContext(ctx, query, positionID).Scan(
		&position.PositionID, &position.AccountID, &position.Symbol, &position.Quantity,
		&position.AvailableQuantity, &position.LockedQuantity, &position.AverageCost,
		&position.MarketValue, &position.Currency, &position.LastUpdated, &position.CreatedAt,
		&position.Metadata,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("position not found: %s", positionID)
	}
	if err != nil {
		r.logger.WithError(err).Error("Failed to get position by ID")
		return nil, fmt.Errorf("failed to get position: %w", err)
	}

	return position, nil
}

func (r *PostgresPositionRepository) GetByAccountAndSymbol(ctx context.Context, accountID, symbol string) (*models.Position, error) {
	query := `
		SELECT position_id, account_id, symbol, quantity, available_quantity, locked_quantity,
			   average_cost, market_value, currency, last_updated, created_at, metadata
		FROM custodian.positions
		WHERE account_id = $1 AND symbol = $2
	`

	position := &models.Position{}
	err := r.db.QueryRowContext(ctx, query, accountID, symbol).Scan(
		&position.PositionID, &position.AccountID, &position.Symbol, &position.Quantity,
		&position.AvailableQuantity, &position.LockedQuantity, &position.AverageCost,
		&position.MarketValue, &position.Currency, &position.LastUpdated, &position.CreatedAt,
		&position.Metadata,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("position not found for account %s and symbol %s", accountID, symbol)
	}
	if err != nil {
		r.logger.WithError(err).Error("Failed to get position by account and symbol")
		return nil, fmt.Errorf("failed to get position: %w", err)
	}

	return position, nil
}

func (r *PostgresPositionRepository) Query(ctx context.Context, query *models.PositionQuery) ([]*models.Position, error) {
	sqlQuery := `
		SELECT position_id, account_id, symbol, quantity, available_quantity, locked_quantity,
			   average_cost, market_value, currency, last_updated, created_at, metadata
		FROM custodian.positions
		WHERE 1=1
	`

	args := []interface{}{}
	argCount := 1

	if query.AccountID != nil {
		sqlQuery += fmt.Sprintf(" AND account_id = $%d", argCount)
		args = append(args, *query.AccountID)
		argCount++
	}

	if query.Symbol != nil {
		sqlQuery += fmt.Sprintf(" AND symbol = $%d", argCount)
		args = append(args, *query.Symbol)
		argCount++
	}

	if query.MinQuantity != nil {
		sqlQuery += fmt.Sprintf(" AND quantity >= $%d", argCount)
		args = append(args, *query.MinQuantity)
		argCount++
	}

	if query.Currency != nil {
		sqlQuery += fmt.Sprintf(" AND currency = $%d", argCount)
		args = append(args, *query.Currency)
		argCount++
	}

	if query.UpdatedAfter != nil {
		sqlQuery += fmt.Sprintf(" AND last_updated > $%d", argCount)
		args = append(args, *query.UpdatedAfter)
		argCount++
	}

	// Add sorting
	if query.SortBy != "" {
		sortOrder := "ASC"
		if strings.ToUpper(query.SortOrder) == "DESC" {
			sortOrder = "DESC"
		}
		sqlQuery += fmt.Sprintf(" ORDER BY %s %s", query.SortBy, sortOrder)
	} else {
		sqlQuery += " ORDER BY last_updated DESC"
	}

	// Add pagination
	if query.Limit > 0 {
		sqlQuery += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, query.Limit)
		argCount++
	}

	if query.Offset > 0 {
		sqlQuery += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, query.Offset)
		argCount++
	}

	rows, err := r.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		r.logger.WithError(err).Error("Failed to query positions")
		return nil, fmt.Errorf("failed to query positions: %w", err)
	}
	defer rows.Close()

	positions := []*models.Position{}
	for rows.Next() {
		position := &models.Position{}
		err := rows.Scan(
			&position.PositionID, &position.AccountID, &position.Symbol, &position.Quantity,
			&position.AvailableQuantity, &position.LockedQuantity, &position.AverageCost,
			&position.MarketValue, &position.Currency, &position.LastUpdated, &position.CreatedAt,
			&position.Metadata,
		)
		if err != nil {
			r.logger.WithError(err).Error("Failed to scan position row")
			return nil, fmt.Errorf("failed to scan position: %w", err)
		}
		positions = append(positions, position)
	}

	return positions, nil
}

func (r *PostgresPositionRepository) Update(ctx context.Context, position *models.Position) error {
	query := `
		UPDATE custodian.positions
		SET account_id = $2, symbol = $3, quantity = $4, available_quantity = $5, locked_quantity = $6,
			average_cost = $7, market_value = $8, currency = $9, last_updated = $10, metadata = $11
		WHERE position_id = $1
	`

	position.LastUpdated = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		position.PositionID, position.AccountID, position.Symbol, position.Quantity,
		position.AvailableQuantity, position.LockedQuantity, position.AverageCost,
		position.MarketValue, position.Currency, position.LastUpdated, position.Metadata,
	)

	if err != nil {
		r.logger.WithError(err).Error("Failed to update position")
		return fmt.Errorf("failed to update position: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("position not found: %s", position.PositionID)
	}

	return nil
}

func (r *PostgresPositionRepository) UpdateAvailableQuantity(ctx context.Context, positionID string, availableQty, lockedQty float64) error {
	query := `
		UPDATE custodian.positions
		SET available_quantity = $2, locked_quantity = $3, last_updated = $4
		WHERE position_id = $1
	`

	result, err := r.db.ExecContext(ctx, query, positionID, availableQty, lockedQty, time.Now())
	if err != nil {
		r.logger.WithError(err).Error("Failed to update available quantity")
		return fmt.Errorf("failed to update available quantity: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("position not found: %s", positionID)
	}

	return nil
}

func (r *PostgresPositionRepository) Delete(ctx context.Context, positionID string) error {
	query := `DELETE FROM custodian.positions WHERE position_id = $1`

	result, err := r.db.ExecContext(ctx, query, positionID)
	if err != nil {
		r.logger.WithError(err).Error("Failed to delete position")
		return fmt.Errorf("failed to delete position: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("position not found: %s", positionID)
	}

	return nil
}

func (r *PostgresPositionRepository) GetByAccount(ctx context.Context, accountID string) ([]*models.Position, error) {
	query := &models.PositionQuery{
		AccountID: &accountID,
		Limit:     1000,
	}
	return r.Query(ctx, query)
}
