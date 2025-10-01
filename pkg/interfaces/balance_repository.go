package interfaces

import (
	"context"
	"github.com/quantfidential/trading-ecosystem/custodian-data-adapter-go/pkg/models"
)

type BalanceRepository interface {
	// Create or update balance
	Upsert(ctx context.Context, balance *models.Balance) error

	// Get balance by ID
	GetByID(ctx context.Context, balanceID string) (*models.Balance, error)

	// Get balance by account and currency
	GetByAccountAndCurrency(ctx context.Context, accountID, currency string) (*models.Balance, error)

	// Query balances with filters
	Query(ctx context.Context, query *models.BalanceQuery) ([]*models.Balance, error)

	// Update available balance (for locking/unlocking)
	UpdateAvailableBalance(ctx context.Context, balanceID string, availableBalance, lockedBalance float64) error

	// Get all balances for account
	GetByAccount(ctx context.Context, accountID string) ([]*models.Balance, error)

	// Atomic balance update (for concurrent operations)
	AtomicUpdate(ctx context.Context, accountID, currency string, availableDelta, lockedDelta float64) error
}
