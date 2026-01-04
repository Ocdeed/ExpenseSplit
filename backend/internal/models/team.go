package models

import (
	"time"

	"github.com/google/uuid"
)

type Team struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedBy uuid.UUID `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

type TeamMember struct {
	TeamID   uuid.UUID `json:"team_id"`
	UserID   uuid.UUID `json:"user_id"`
	Role     string    `json:"role"` // "admin", "member"
	JoinedAt time.Time `json:"joined_at"`
}

type TeamCreateRequest struct {
	Name string `json:"name"`
}

type TeamResponse struct {
	ID        uuid.UUID      `json:"id"`
	Name      string         `json:"name"`
	CreatedBy uuid.UUID      `json:"created_by"`
	CreatedAt time.Time      `json:"created_at"`
	Members   []MemberDetail `json:"members,omitempty"`
}

type MemberDetail struct {
	UserID   uuid.UUID `json:"user_id"`
	Email    string    `json:"email"`
	Name     string    `json:"name"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

type AddMemberRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

func (t *Team) ToResponse() TeamResponse {
	return TeamResponse{
		ID:        t.ID,
		Name:      t.Name,
		CreatedBy: t.CreatedBy,
		CreatedAt: t.CreatedAt,
	}
}
