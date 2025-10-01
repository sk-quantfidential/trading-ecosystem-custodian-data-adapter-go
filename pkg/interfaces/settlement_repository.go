package interfaces

import (
	"context"
	"github.com/quantfidential/trading-ecosystem/custodian-data-adapter-go/pkg/models"
)

type SettlementRepository interface {
	// Create a new settlement instruction
	Create(ctx context.Context, settlement *models.Settlement) error

	// Get settlement by ID
	GetByID(ctx context.Context, settlementID string) (*models.Settlement, error)

	// Get settlement by external ID
	GetByExternalID(ctx context.Context, externalID string) (*models.Settlement, error)

	// Query settlements with filters
	Query(ctx context.Context, query *models.SettlementQuery) ([]*models.Settlement, error)

	// Update settlement status
	UpdateStatus(ctx context.Context, settlementID string, status models.SettlementStatus) error

	// Complete settlement
	Complete(ctx context.Context, settlementID string) error

	// Cancel settlement
	Cancel(ctx context.Context, settlementID string) error

	// Get pending settlements for account
	GetPendingByAccount(ctx context.Context, accountID string) ([]*models.Settlement, error)
}
