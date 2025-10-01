package adapters

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/quantfidential/trading-ecosystem/custodian-data-adapter-go/pkg/interfaces"
	"github.com/quantfidential/trading-ecosystem/custodian-data-adapter-go/pkg/models"
	"github.com/sirupsen/logrus"
)

type PostgresSettlementRepository struct {
	db     *sql.DB
	logger *logrus.Logger
}

func NewPostgresSettlementRepository(db *sql.DB, logger *logrus.Logger) interfaces.SettlementRepository {
	return &PostgresSettlementRepository{
		db:     db,
		logger: logger,
	}
}

func (r *PostgresSettlementRepository) Create(ctx context.Context, settlement *models.Settlement) error {
	query := `
		INSERT INTO custodian.settlements (
			settlement_id, external_id, settlement_type, account_id, symbol, quantity, status,
			source_account, destination_account, initiated_at, completed_at, expected_settlement_date, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := r.db.ExecContext(ctx, query,
		settlement.SettlementID, settlement.ExternalID, settlement.SettlementType, settlement.AccountID,
		settlement.Symbol, settlement.Quantity, settlement.Status, settlement.SourceAccount,
		settlement.DestinationAccount, settlement.InitiatedAt, settlement.CompletedAt,
		settlement.ExpectedSettlementDate, settlement.Metadata,
	)

	if err != nil {
		r.logger.WithError(err).Error("Failed to create settlement")
		return fmt.Errorf("failed to create settlement: %w", err)
	}

	return nil
}

func (r *PostgresSettlementRepository) GetByID(ctx context.Context, settlementID string) (*models.Settlement, error) {
	query := `
		SELECT settlement_id, external_id, settlement_type, account_id, symbol, quantity, status,
			   source_account, destination_account, initiated_at, completed_at, expected_settlement_date, metadata
		FROM custodian.settlements
		WHERE settlement_id = $1
	`

	settlement := &models.Settlement{}
	err := r.db.QueryRowContext(ctx, query, settlementID).Scan(
		&settlement.SettlementID, &settlement.ExternalID, &settlement.SettlementType, &settlement.AccountID,
		&settlement.Symbol, &settlement.Quantity, &settlement.Status, &settlement.SourceAccount,
		&settlement.DestinationAccount, &settlement.InitiatedAt, &settlement.CompletedAt,
		&settlement.ExpectedSettlementDate, &settlement.Metadata,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("settlement not found: %s", settlementID)
	}
	if err != nil {
		r.logger.WithError(err).Error("Failed to get settlement")
		return nil, fmt.Errorf("failed to get settlement: %w", err)
	}

	return settlement, nil
}

func (r *PostgresSettlementRepository) GetByExternalID(ctx context.Context, externalID string) (*models.Settlement, error) {
	query := `
		SELECT settlement_id, external_id, settlement_type, account_id, symbol, quantity, status,
			   source_account, destination_account, initiated_at, completed_at, expected_settlement_date, metadata
		FROM custodian.settlements
		WHERE external_id = $1
	`

	settlement := &models.Settlement{}
	err := r.db.QueryRowContext(ctx, query, externalID).Scan(
		&settlement.SettlementID, &settlement.ExternalID, &settlement.SettlementType, &settlement.AccountID,
		&settlement.Symbol, &settlement.Quantity, &settlement.Status, &settlement.SourceAccount,
		&settlement.DestinationAccount, &settlement.InitiatedAt, &settlement.CompletedAt,
		&settlement.ExpectedSettlementDate, &settlement.Metadata,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("settlement not found with external ID: %s", externalID)
	}
	if err != nil {
		r.logger.WithError(err).Error("Failed to get settlement by external ID")
		return nil, fmt.Errorf("failed to get settlement: %w", err)
	}

	return settlement, nil
}

func (r *PostgresSettlementRepository) Query(ctx context.Context, query *models.SettlementQuery) ([]*models.Settlement, error) {
	// Implementation similar to Position Query (simplified for brevity)
	sqlQuery := `
		SELECT settlement_id, external_id, settlement_type, account_id, symbol, quantity, status,
			   source_account, destination_account, initiated_at, completed_at, expected_settlement_date, metadata
		FROM custodian.settlements
		WHERE 1=1
	`

	args := []interface{}{}
	argCount := 1

	if query.AccountID != nil {
		sqlQuery += fmt.Sprintf(" AND account_id = $%d", argCount)
		args = append(args, *query.AccountID)
		argCount++
	}

	if query.Status != nil {
		sqlQuery += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, *query.Status)
		argCount++
	}

	if query.SettlementType != nil {
		sqlQuery += fmt.Sprintf(" AND settlement_type = $%d", argCount)
		args = append(args, *query.SettlementType)
		argCount++
	}

	sqlQuery += " ORDER BY initiated_at DESC"

	if query.Limit > 0 {
		sqlQuery += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, query.Limit)
		argCount++
	}

	rows, err := r.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		r.logger.WithError(err).Error("Failed to query settlements")
		return nil, fmt.Errorf("failed to query settlements: %w", err)
	}
	defer rows.Close()

	settlements := []*models.Settlement{}
	for rows.Next() {
		settlement := &models.Settlement{}
		err := rows.Scan(
			&settlement.SettlementID, &settlement.ExternalID, &settlement.SettlementType, &settlement.AccountID,
			&settlement.Symbol, &settlement.Quantity, &settlement.Status, &settlement.SourceAccount,
			&settlement.DestinationAccount, &settlement.InitiatedAt, &settlement.CompletedAt,
			&settlement.ExpectedSettlementDate, &settlement.Metadata,
		)
		if err != nil {
			r.logger.WithError(err).Error("Failed to scan settlement")
			return nil, fmt.Errorf("failed to scan settlement: %w", err)
		}
		settlements = append(settlements, settlement)
	}

	return settlements, nil
}

func (r *PostgresSettlementRepository) UpdateStatus(ctx context.Context, settlementID string, status models.SettlementStatus) error {
	query := `UPDATE custodian.settlements SET status = $2 WHERE settlement_id = $1`

	result, err := r.db.ExecContext(ctx, query, settlementID, status)
	if err != nil {
		r.logger.WithError(err).Error("Failed to update settlement status")
		return fmt.Errorf("failed to update settlement status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("settlement not found: %s", settlementID)
	}

	return nil
}

func (r *PostgresSettlementRepository) Complete(ctx context.Context, settlementID string) error {
	now := time.Now()
	query := `UPDATE custodian.settlements SET status = $2, completed_at = $3 WHERE settlement_id = $1`

	result, err := r.db.ExecContext(ctx, query, settlementID, models.SettlementStatusCompleted, now)
	if err != nil {
		r.logger.WithError(err).Error("Failed to complete settlement")
		return fmt.Errorf("failed to complete settlement: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("settlement not found: %s", settlementID)
	}

	return nil
}

func (r *PostgresSettlementRepository) Cancel(ctx context.Context, settlementID string) error {
	query := `UPDATE custodian.settlements SET status = $2 WHERE settlement_id = $1`

	result, err := r.db.ExecContext(ctx, query, settlementID, models.SettlementStatusCancelled)
	if err != nil {
		r.logger.WithError(err).Error("Failed to cancel settlement")
		return fmt.Errorf("failed to cancel settlement: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("settlement not found: %s", settlementID)
	}

	return nil
}

func (r *PostgresSettlementRepository) GetPendingByAccount(ctx context.Context, accountID string) ([]*models.Settlement, error) {
	status := models.SettlementStatusPending
	query := &models.SettlementQuery{
		AccountID: &accountID,
		Status:    &status,
		Limit:     1000,
	}
	return r.Query(ctx, query)
}
