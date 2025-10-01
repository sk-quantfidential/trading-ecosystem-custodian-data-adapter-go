package models

import (
	"encoding/json"
	"time"
)

type Balance struct {
	BalanceID        string          `json:"balance_id" db:"balance_id"`
	AccountID        string          `json:"account_id" db:"account_id"`
	Currency         string          `json:"currency" db:"currency"`
	AvailableBalance float64         `json:"available_balance" db:"available_balance"`
	LockedBalance    float64         `json:"locked_balance" db:"locked_balance"`
	TotalBalance     float64         `json:"total_balance" db:"total_balance"`
	LastUpdated      time.Time       `json:"last_updated" db:"last_updated"`
	Metadata         json.RawMessage `json:"metadata,omitempty" db:"metadata"`
}

type BalanceQuery struct {
	AccountID    *string
	Currency     *string
	MinBalance   *float64
	UpdatedAfter *time.Time
	Limit        int
	Offset       int
	SortBy       string
	SortOrder    string
}
