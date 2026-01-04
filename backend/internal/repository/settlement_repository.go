package repository

import (
	"time"

	"github.com/expensesplit/backend/internal/database"
	"github.com/expensesplit/backend/internal/models"
	"github.com/google/uuid"
)

type SettlementRepository struct {
	db *database.DB
}

func NewSettlementRepository(db *database.DB) *SettlementRepository {
	return &SettlementRepository{db: db}
}

func (r *SettlementRepository) Create(settlement *models.Settlement) error {
	settlement.ID = uuid.New()
	settlement.CreatedAt = time.Now()

	query := `
		INSERT INTO settlements (id, team_id, from_user, to_user, amount, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(query, settlement.ID, settlement.TeamID, settlement.FromUser,
		settlement.ToUser, settlement.Amount, settlement.CreatedAt)
	return err
}

func (r *SettlementRepository) GetByTeamID(teamID uuid.UUID) ([]models.Settlement, error) {
	query := `
		SELECT id, team_id, from_user, to_user, amount, created_at
		FROM settlements WHERE team_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settlements []models.Settlement
	for rows.Next() {
		settlement := models.Settlement{}
		err := rows.Scan(&settlement.ID, &settlement.TeamID, &settlement.FromUser,
			&settlement.ToUser, &settlement.Amount, &settlement.CreatedAt)
		if err != nil {
			return nil, err
		}
		settlements = append(settlements, settlement)
	}
	return settlements, nil
}

func (r *SettlementRepository) GetByUsers(teamID, fromUser, toUser uuid.UUID) ([]models.Settlement, error) {
	query := `
		SELECT id, team_id, from_user, to_user, amount, created_at
		FROM settlements 
		WHERE team_id = $1 AND ((from_user = $2 AND to_user = $3) OR (from_user = $3 AND to_user = $2))
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query, teamID, fromUser, toUser)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settlements []models.Settlement
	for rows.Next() {
		settlement := models.Settlement{}
		err := rows.Scan(&settlement.ID, &settlement.TeamID, &settlement.FromUser,
			&settlement.ToUser, &settlement.Amount, &settlement.CreatedAt)
		if err != nil {
			return nil, err
		}
		settlements = append(settlements, settlement)
	}
	return settlements, nil
}
