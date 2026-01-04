package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/expensesplit/backend/internal/database"
	"github.com/expensesplit/backend/internal/models"
	"github.com/google/uuid"
)

var (
	ErrTeamNotFound     = errors.New("team not found")
	ErrNotTeamMember    = errors.New("user is not a member of this team")
	ErrAlreadyMember    = errors.New("user is already a member of this team")
	ErrCannotRemoveOwner = errors.New("cannot remove team owner")
)

type TeamRepository struct {
	db *database.DB
}

func NewTeamRepository(db *database.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

func (r *TeamRepository) Create(team *models.Team, creatorID uuid.UUID) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	team.ID = uuid.New()
	team.CreatedBy = creatorID
	team.CreatedAt = time.Now()

	// Create team
	query := `INSERT INTO teams (id, name, created_by, created_at) VALUES ($1, $2, $3, $4)`
	_, err = tx.Exec(query, team.ID, team.Name, team.CreatedBy, team.CreatedAt)
	if err != nil {
		return err
	}

	// Add creator as admin member
	memberQuery := `INSERT INTO team_members (team_id, user_id, role, joined_at) VALUES ($1, $2, $3, $4)`
	_, err = tx.Exec(memberQuery, team.ID, creatorID, "admin", time.Now())
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *TeamRepository) GetByID(id uuid.UUID) (*models.Team, error) {
	team := &models.Team{}
	query := `SELECT id, name, created_by, created_at FROM teams WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(&team.ID, &team.Name, &team.CreatedBy, &team.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrTeamNotFound
	}
	if err != nil {
		return nil, err
	}
	return team, nil
}

func (r *TeamRepository) GetUserTeams(userID uuid.UUID) ([]*models.Team, error) {
	query := `
		SELECT t.id, t.name, t.created_by, t.created_at
		FROM teams t
		INNER JOIN team_members tm ON t.id = tm.team_id
		WHERE tm.user_id = $1
		ORDER BY t.created_at DESC
	`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []*models.Team
	for rows.Next() {
		team := &models.Team{}
		err := rows.Scan(&team.ID, &team.Name, &team.CreatedBy, &team.CreatedAt)
		if err != nil {
			return nil, err
		}
		teams = append(teams, team)
	}
	return teams, nil
}

func (r *TeamRepository) GetTeamMembers(teamID uuid.UUID) ([]models.MemberDetail, error) {
	query := `
		SELECT u.id, u.email, u.name, tm.role, tm.joined_at
		FROM users u
		INNER JOIN team_members tm ON u.id = tm.user_id
		WHERE tm.team_id = $1
		ORDER BY tm.joined_at ASC
	`
	rows, err := r.db.Query(query, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []models.MemberDetail
	for rows.Next() {
		member := models.MemberDetail{}
		err := rows.Scan(&member.UserID, &member.Email, &member.Name, &member.Role, &member.JoinedAt)
		if err != nil {
			return nil, err
		}
		members = append(members, member)
	}
	return members, nil
}

func (r *TeamRepository) AddMember(teamID, userID uuid.UUID, role string) error {
	// Check if already a member
	var exists bool
	err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM team_members WHERE team_id = $1 AND user_id = $2)", teamID, userID).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return ErrAlreadyMember
	}

	query := `INSERT INTO team_members (team_id, user_id, role, joined_at) VALUES ($1, $2, $3, $4)`
	_, err = r.db.Exec(query, teamID, userID, role, time.Now())
	return err
}

func (r *TeamRepository) RemoveMember(teamID, userID uuid.UUID) error {
	// Check if user is the owner
	var createdBy uuid.UUID
	err := r.db.QueryRow("SELECT created_by FROM teams WHERE id = $1", teamID).Scan(&createdBy)
	if err != nil {
		return err
	}
	if createdBy == userID {
		return ErrCannotRemoveOwner
	}

	query := `DELETE FROM team_members WHERE team_id = $1 AND user_id = $2`
	result, err := r.db.Exec(query, teamID, userID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotTeamMember
	}
	return nil
}

func (r *TeamRepository) IsMember(teamID, userID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM team_members WHERE team_id = $1 AND user_id = $2)", teamID, userID).Scan(&exists)
	return exists, err
}

func (r *TeamRepository) IsAdmin(teamID, userID uuid.UUID) (bool, error) {
	var role string
	err := r.db.QueryRow("SELECT role FROM team_members WHERE team_id = $1 AND user_id = $2", teamID, userID).Scan(&role)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return role == "admin", nil
}

func (r *TeamRepository) UpdateMemberRole(teamID, userID uuid.UUID, role string) error {
	query := `UPDATE team_members SET role = $1 WHERE team_id = $2 AND user_id = $3`
	result, err := r.db.Exec(query, role, teamID, userID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotTeamMember
	}
	return nil
}

func (r *TeamRepository) Update(team *models.Team) error {
	query := `UPDATE teams SET name = $1 WHERE id = $2`
	result, err := r.db.Exec(query, team.Name, team.ID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrTeamNotFound
	}
	return nil
}

func (r *TeamRepository) Delete(teamID uuid.UUID) error {
	query := `DELETE FROM teams WHERE id = $1`
	result, err := r.db.Exec(query, teamID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrTeamNotFound
	}
	return nil
}
