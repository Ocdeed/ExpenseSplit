package models

import (
	"time"

	"github.com/google/uuid"
)

type ApprovalStatus string

const (
	ApprovalStatusPending  ApprovalStatus = "pending"
	ApprovalStatusApproved ApprovalStatus = "approved"
	ApprovalStatusRejected ApprovalStatus = "rejected"
)

type Approval struct {
	ID         uuid.UUID      `json:"id"`
	ExpenseID  uuid.UUID      `json:"expense_id"`
	ApprovedBy uuid.UUID      `json:"approved_by,omitempty"`
	Status     ApprovalStatus `json:"status"`
	Comment    string         `json:"comment,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	ApprovedAt *time.Time     `json:"approved_at,omitempty"`
}

type ApprovalRequest struct {
	Status  ApprovalStatus `json:"status"`
	Comment string         `json:"comment,omitempty"`
}

type ApprovalResponse struct {
	ID         uuid.UUID       `json:"id"`
	Expense    ExpenseResponse `json:"expense"`
	ApprovedBy *UserResponse   `json:"approved_by,omitempty"`
	Status     ApprovalStatus  `json:"status"`
	Comment    string          `json:"comment,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
	ApprovedAt *time.Time      `json:"approved_at,omitempty"`
}

type ReimbursementSummary struct {
	TeamID          uuid.UUID           `json:"team_id"`
	TeamName        string              `json:"team_name"`
	Period          string              `json:"period"`
	TotalExpenses   float64             `json:"total_expenses"`
	TotalApproved   float64             `json:"total_approved"`
	TotalPending    float64             `json:"total_pending"`
	TotalRejected   float64             `json:"total_rejected"`
	Expenses        []ExpenseResponse   `json:"expenses"`
	Settlements     []SettlementSummary `json:"settlements"`
	GeneratedAt     time.Time           `json:"generated_at"`
}

type SettlementSummary struct {
	FromUser UserResponse `json:"from_user"`
	ToUser   UserResponse `json:"to_user"`
	Amount   float64      `json:"amount"`
}
