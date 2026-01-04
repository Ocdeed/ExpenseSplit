package models

import (
	"time"

	"github.com/google/uuid"
)

type Balance struct {
	ID        uuid.UUID `json:"id"`
	TeamID    uuid.UUID `json:"team_id"`
	FromUser  uuid.UUID `json:"from_user"`
	ToUser    uuid.UUID `json:"to_user"`
	Amount    float64   `json:"amount"`
	UpdatedAt time.Time `json:"updated_at"`
}

type BalanceResponse struct {
	FromUser UserResponse `json:"from_user"`
	ToUser   UserResponse `json:"to_user"`
	Amount   float64      `json:"amount"`
}

type UserBalanceSummary struct {
	User       UserResponse `json:"user"`
	TotalOwed  float64      `json:"total_owed"`  // Amount this user owes to others
	TotalOwing float64      `json:"total_owing"` // Amount others owe to this user
	NetBalance float64      `json:"net_balance"` // Positive = others owe you, Negative = you owe others
}

type TeamBalanceSummary struct {
	TeamID   uuid.UUID            `json:"team_id"`
	TeamName string               `json:"team_name"`
	Balances []BalanceResponse    `json:"balances"`
	Members  []UserBalanceSummary `json:"members"`
}

type SettlementRequest struct {
	FromUser uuid.UUID `json:"from_user"`
	ToUser   uuid.UUID `json:"to_user"`
	Amount   float64   `json:"amount"`
}

type Settlement struct {
	ID        uuid.UUID `json:"id"`
	TeamID    uuid.UUID `json:"team_id"`
	FromUser  uuid.UUID `json:"from_user"`
	ToUser    uuid.UUID `json:"to_user"`
	Amount    float64   `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}
