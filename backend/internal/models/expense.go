package models

import (
	"time"

	"github.com/google/uuid"
)

type SplitType string

const (
	SplitTypeEqual   SplitType = "equal"
	SplitTypeCustom  SplitType = "custom"
	SplitTypePercent SplitType = "percent"
)

type Expense struct {
	ID          uuid.UUID `json:"id"`
	TeamID      uuid.UUID `json:"team_id"`
	PaidBy      uuid.UUID `json:"paid_by"`
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	ReceiptURL  string    `json:"receipt_url,omitempty"`
	SplitType   SplitType `json:"split_type"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ExpenseSplit struct {
	ID        uuid.UUID `json:"id"`
	ExpenseID uuid.UUID `json:"expense_id"`
	UserID    uuid.UUID `json:"user_id"`
	Amount    float64   `json:"amount"`
	Percent   float64   `json:"percent,omitempty"`
	IsSettled bool      `json:"is_settled"`
}

type ExpenseCreateRequest struct {
	Amount      float64            `json:"amount"`
	Description string             `json:"description"`
	Category    string             `json:"category"`
	SplitType   SplitType          `json:"split_type"`
	SplitWith   []uuid.UUID        `json:"split_with"`             // User IDs to split with
	CustomSplit []CustomSplitEntry `json:"custom_split,omitempty"` // For custom splits
}

type CustomSplitEntry struct {
	UserID  uuid.UUID `json:"user_id"`
	Amount  float64   `json:"amount,omitempty"`
	Percent float64   `json:"percent,omitempty"`
}

type ExpenseUpdateRequest struct {
	Amount      *float64 `json:"amount,omitempty"`
	Description *string  `json:"description,omitempty"`
	Category    *string  `json:"category,omitempty"`
}

type ExpenseResponse struct {
	ID             uuid.UUID            `json:"id"`
	TeamID         uuid.UUID            `json:"team_id"`
	PaidBy         UserResponse         `json:"paid_by"`
	Amount         float64              `json:"amount"`
	Description    string               `json:"description"`
	Category       string               `json:"category"`
	ReceiptURL     string               `json:"receipt_url,omitempty"`
	SplitType      SplitType            `json:"split_type"`
	Splits         []ExpenseSplitDetail `json:"splits"`
	ApprovalStatus ApprovalStatus       `json:"approval_status"`
	CreatedAt      time.Time            `json:"created_at"`
}

type ExpenseSplitDetail struct {
	ID        uuid.UUID    `json:"id"`
	User      UserResponse `json:"user"`
	Amount    float64      `json:"amount"`
	Percent   float64      `json:"percent,omitempty"`
	IsSettled bool         `json:"is_settled"`
}

// Categories for expenses
var ExpenseCategories = []string{
	"Food & Dining",
	"Transportation",
	"Accommodation",
	"Office Supplies",
	"Software & Tools",
	"Entertainment",
	"Travel",
	"Utilities",
	"Marketing",
	"Other",
}
