package models

import (
	"encoding/json"
	"time"
)

type Position struct {
	PositionID        string          `json:"position_id" db:"position_id"`
	AccountID         string          `json:"account_id" db:"account_id"`
	Symbol            string          `json:"symbol" db:"symbol"`
	Quantity          float64         `json:"quantity" db:"quantity"`
	AvailableQuantity float64         `json:"available_quantity" db:"available_quantity"`
	LockedQuantity    float64         `json:"locked_quantity" db:"locked_quantity"`
	AverageCost       *float64        `json:"average_cost,omitempty" db:"average_cost"`
	MarketValue       *float64        `json:"market_value,omitempty" db:"market_value"`
	Currency          string          `json:"currency" db:"currency"`
	LastUpdated       time.Time       `json:"last_updated" db:"last_updated"`
	CreatedAt         time.Time       `json:"created_at" db:"created_at"`
	Metadata          json.RawMessage `json:"metadata,omitempty" db:"metadata"`
}

type PositionQuery struct {
	AccountID    *string
	Symbol       *string
	MinQuantity  *float64
	Currency     *string
	UpdatedAfter *time.Time
	Limit        int
	Offset       int
	SortBy       string
	SortOrder    string
}
