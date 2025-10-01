package models

import (
	"encoding/json"
	"time"
)

type SettlementType string

const (
	SettlementTypeDeposit    SettlementType = "DEPOSIT"
	SettlementTypeWithdrawal SettlementType = "WITHDRAWAL"
	SettlementTypeTransfer   SettlementType = "TRANSFER"
)

type SettlementStatus string

const (
	SettlementStatusPending    SettlementStatus = "PENDING"
	SettlementStatusInProgress SettlementStatus = "IN_PROGRESS"
	SettlementStatusCompleted  SettlementStatus = "COMPLETED"
	SettlementStatusFailed     SettlementStatus = "FAILED"
	SettlementStatusCancelled  SettlementStatus = "CANCELLED"
)

type Settlement struct {
	SettlementID            string           `json:"settlement_id" db:"settlement_id"`
	ExternalID              *string          `json:"external_id,omitempty" db:"external_id"`
	SettlementType          SettlementType   `json:"settlement_type" db:"settlement_type"`
	AccountID               string           `json:"account_id" db:"account_id"`
	Symbol                  string           `json:"symbol" db:"symbol"`
	Quantity                float64          `json:"quantity" db:"quantity"`
	Status                  SettlementStatus `json:"status" db:"status"`
	SourceAccount           *string          `json:"source_account,omitempty" db:"source_account"`
	DestinationAccount      *string          `json:"destination_account,omitempty" db:"destination_account"`
	InitiatedAt             time.Time        `json:"initiated_at" db:"initiated_at"`
	CompletedAt             *time.Time       `json:"completed_at,omitempty" db:"completed_at"`
	ExpectedSettlementDate  *time.Time       `json:"expected_settlement_date,omitempty" db:"expected_settlement_date"`
	Metadata                json.RawMessage  `json:"metadata,omitempty" db:"metadata"`
}

type SettlementQuery struct {
	AccountID      *string
	Status         *SettlementStatus
	SettlementType *SettlementType
	Symbol         *string
	InitiatedAfter *time.Time
	Limit          int
	Offset         int
	SortBy         string
	SortOrder      string
}
