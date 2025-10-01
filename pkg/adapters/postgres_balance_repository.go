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

type PostgresBalanceRepository struct {
	db     *sql.DB
	logger *logrus.Logger
}

func NewPostgresBalanceRepository(db *sql.DB, logger *logrus.Logger) interfaces.BalanceRepository {
	return &PostgresBalanceRepository{
		db:     db,
		logger: logger,
	}
}

func (r *PostgresBalanceRepository) Upsert(ctx context.Context, balance *models.Balance) error {
	query := `
		INSERT INTO custodian.balances (
			balance_id, account_id, currency, available_balance, locked_balance, total_balance, last_updated, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (account_id, currency)
		DO UPDATE SET
			available_balance = EXCLUDED.available_balance,
			locked_balance = EXCLUDED.locked_balance,
			total_balance = EXCLUDED.total_balance,
			last_updated = EXCLUDED.last_updated,
			metadata = EXCLUDED.metadata
	`

	balance.LastUpdated = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		balance.BalanceID, balance.AccountID, balance.Currency, balance.AvailableBalance,
		balance.LockedBalance, balance.TotalBalance, balance.LastUpdated, balance.Metadata,
	)

	if err != nil {
		r.logger.WithError(err).Error("Failed to upsert balance")
		return fmt.Errorf("failed to upsert balance: %w", err)
	}

	return nil
}

func (r *PostgresBalanceRepository) GetByID(ctx context.Context, balanceID string) (*models.Balance, error) {
	query := `
		SELECT balance_id, account_id, currency, available_balance, locked_balance, total_balance, last_updated, metadata
		FROM custodian.balances
		WHERE balance_id = $1
	`

	balance := &models.Balance{}
	err := r.db.QueryRowContext(ctx, query, balanceID).Scan(
		&balance.BalanceID, &balance.AccountID, &balance.Currency, &balance.AvailableBalance,
		&balance.LockedBalance, &balance.TotalBalance, &balance.LastUpdated, &balance.Metadata,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("balance not found: %s", balanceID)
	}
	if err != nil {
		r.logger.WithError(err).Error("Failed to get balance")
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	return balance, nil
}

func (r *PostgresBalanceRepository) GetByAccountAndCurrency(ctx context.Context, accountID, currency string) (*models.Balance, error) {
	query := `
		SELECT balance_id, account_id, currency, available_balance, locked_balance, total_balance, last_updated, metadata
		FROM custodian.balances
		WHERE account_id = $1 AND currency = $2
	`

	balance := &models.Balance{}
	err := r.db.QueryRowContext(ctx, query, accountID, currency).Scan(
		&balance.BalanceID, &balance.AccountID, &balance.Currency, &balance.AvailableBalance,
		&balance.LockedBalance, &balance.TotalBalance, &balance.LastUpdated, &balance.Metadata,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("balance not found for account %s and currency %s", accountID, currency)
	}
	if err != nil {
		r.logger.WithError(err).Error("Failed to get balance")
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	return balance, nil
}

func (r *PostgresBalanceRepository) Query(ctx context.Context, query *models.BalanceQuery) ([]*models.Balance, error) {
	sqlQuery := `
		SELECT balance_id, account_id, currency, available_balance, locked_balance, total_balance, last_updated, metadata
		FROM custodian.balances
		WHERE 1=1
	`

	args := []interface{}{}
	argCount := 1

	if query.AccountID != nil {
		sqlQuery += fmt.Sprintf(" AND account_id = $%d", argCount)
		args = append(args, *query.AccountID)
		argCount++
	}

	if query.Currency != nil {
		sqlQuery += fmt.Sprintf(" AND currency = $%d", argCount)
		args = append(args, *query.Currency)
		argCount++
	}

	sqlQuery += " ORDER BY last_updated DESC"

	if query.Limit > 0 {
		sqlQuery += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, query.Limit)
		argCount++
	}

	rows, err := r.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		r.logger.WithError(err).Error("Failed to query balances")
		return nil, fmt.Errorf("failed to query balances: %w", err)
	}
	defer rows.Close()

	balances := []*models.Balance{}
	for rows.Next() {
		balance := &models.Balance{}
		err := rows.Scan(
			&balance.BalanceID, &balance.AccountID, &balance.Currency, &balance.AvailableBalance,
			&balance.LockedBalance, &balance.TotalBalance, &balance.LastUpdated, &balance.Metadata,
		)
		if err != nil {
			r.logger.WithError(err).Error("Failed to scan balance")
			return nil, fmt.Errorf("failed to scan balance: %w", err)
		}
		balances = append(balances, balance)
	}

	return balances, nil
}

func (r *PostgresBalanceRepository) UpdateAvailableBalance(ctx context.Context, balanceID string, availableBalance, lockedBalance float64) error {
	query := `
		UPDATE custodian.balances
		SET available_balance = $2, locked_balance = $3, total_balance = $2 + $3, last_updated = $4
		WHERE balance_id = $1
	`

	result, err := r.db.ExecContext(ctx, query, balanceID, availableBalance, lockedBalance, time.Now())
	if err != nil {
		r.logger.WithError(err).Error("Failed to update available balance")
		return fmt.Errorf("failed to update available balance: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("balance not found: %s", balanceID)
	}

	return nil
}

func (r *PostgresBalanceRepository) GetByAccount(ctx context.Context, accountID string) ([]*models.Balance, error) {
	query := &models.BalanceQuery{
		AccountID: &accountID,
		Limit:     1000,
	}
	return r.Query(ctx, query)
}

func (r *PostgresBalanceRepository) AtomicUpdate(ctx context.Context, accountID, currency string, availableDelta, lockedDelta float64) error {
	query := `
		UPDATE custodian.balances
		SET available_balance = available_balance + $3,
			locked_balance = locked_balance + $4,
			total_balance = available_balance + $3 + locked_balance + $4,
			last_updated = $5
		WHERE account_id = $1 AND currency = $2
	`

	result, err := r.db.ExecContext(ctx, query, accountID, currency, availableDelta, lockedDelta, time.Now())
	if err != nil {
		r.logger.WithError(err).Error("Failed to atomic update balance")
		return fmt.Errorf("failed to atomic update balance: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("balance not found for account %s and currency %s", accountID, currency)
	}

	return nil
}
