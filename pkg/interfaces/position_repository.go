package interfaces

import (
	"context"
	"github.com/quantfidential/trading-ecosystem/custodian-data-adapter-go/pkg/models"
)

type PositionRepository interface {
	// Create a new position
	Create(ctx context.Context, position *models.Position) error

	// Get position by ID
	GetByID(ctx context.Context, positionID string) (*models.Position, error)

	// Get position by account and symbol
	GetByAccountAndSymbol(ctx context.Context, accountID, symbol string) (*models.Position, error)

	// Query positions with filters
	Query(ctx context.Context, query *models.PositionQuery) ([]*models.Position, error)

	// Update position
	Update(ctx context.Context, position *models.Position) error

	// Update available quantity (for locking/unlocking)
	UpdateAvailableQuantity(ctx context.Context, positionID string, availableQty, lockedQty float64) error

	// Delete position
	Delete(ctx context.Context, positionID string) error

	// Get positions by account
	GetByAccount(ctx context.Context, accountID string) ([]*models.Position, error)
}
