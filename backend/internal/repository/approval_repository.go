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
	ErrApprovalNotFound = errors.New("approval not found")
)

type ApprovalRepository struct {
	db *database.DB
}

func NewApprovalRepository(db *database.DB) *ApprovalRepository {
	return &ApprovalRepository{db: db}
}

func (r *ApprovalRepository) Create(approval *models.Approval) error {
	approval.ID = uuid.New()
	approval.CreatedAt = time.Now()
	approval.Status = models.ApprovalStatusPending

	query := `
		INSERT INTO approvals (id, expense_id, status, comment, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Exec(query, approval.ID, approval.ExpenseID, approval.Status, approval.Comment, approval.CreatedAt)
	return err
}

func (r *ApprovalRepository) GetByID(id uuid.UUID) (*models.Approval, error) {
	approval := &models.Approval{}
	query := `
		SELECT id, expense_id, approved_by, status, comment, created_at, approved_at
		FROM approvals WHERE id = $1
	`
	var approvedBy sql.NullString
	var approvedAt sql.NullTime
	err := r.db.QueryRow(query, id).Scan(
		&approval.ID, &approval.ExpenseID, &approvedBy, &approval.Status,
		&approval.Comment, &approval.CreatedAt, &approvedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrApprovalNotFound
	}
	if err != nil {
		return nil, err
	}
	if approvedBy.Valid {
		uid, _ := uuid.Parse(approvedBy.String)
		approval.ApprovedBy = uid
	}
	if approvedAt.Valid {
		approval.ApprovedAt = &approvedAt.Time
	}
	return approval, nil
}

func (r *ApprovalRepository) GetByExpenseID(expenseID uuid.UUID) (*models.Approval, error) {
	approval := &models.Approval{}
	query := `
		SELECT id, expense_id, approved_by, status, comment, created_at, approved_at
		FROM approvals WHERE expense_id = $1
	`
	var approvedBy sql.NullString
	var approvedAt sql.NullTime
	err := r.db.QueryRow(query, expenseID).Scan(
		&approval.ID, &approval.ExpenseID, &approvedBy, &approval.Status,
		&approval.Comment, &approval.CreatedAt, &approvedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrApprovalNotFound
	}
	if err != nil {
		return nil, err
	}
	if approvedBy.Valid {
		uid, _ := uuid.Parse(approvedBy.String)
		approval.ApprovedBy = uid
	}
	if approvedAt.Valid {
		approval.ApprovedAt = &approvedAt.Time
	}
	return approval, nil
}

func (r *ApprovalRepository) GetPendingByTeamID(teamID uuid.UUID) ([]models.Approval, error) {
	query := `
		SELECT a.id, a.expense_id, a.approved_by, a.status, a.comment, a.created_at, a.approved_at
		FROM approvals a
		INNER JOIN expenses e ON a.expense_id = e.id
		WHERE e.team_id = $1 AND a.status = 'pending'
		ORDER BY a.created_at DESC
	`
	rows, err := r.db.Query(query, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var approvals []models.Approval
	for rows.Next() {
		approval := models.Approval{}
		var approvedBy sql.NullString
		var approvedAt sql.NullTime
		err := rows.Scan(
			&approval.ID, &approval.ExpenseID, &approvedBy, &approval.Status,
			&approval.Comment, &approval.CreatedAt, &approvedAt,
		)
		if err != nil {
			return nil, err
		}
		if approvedBy.Valid {
			uid, _ := uuid.Parse(approvedBy.String)
			approval.ApprovedBy = uid
		}
		if approvedAt.Valid {
			approval.ApprovedAt = &approvedAt.Time
		}
		approvals = append(approvals, approval)
	}
	return approvals, nil
}

func (r *ApprovalRepository) GetAllByTeamID(teamID uuid.UUID) ([]models.Approval, error) {
	query := `
		SELECT a.id, a.expense_id, a.approved_by, a.status, a.comment, a.created_at, a.approved_at
		FROM approvals a
		INNER JOIN expenses e ON a.expense_id = e.id
		WHERE e.team_id = $1
		ORDER BY a.created_at DESC
	`
	rows, err := r.db.Query(query, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var approvals []models.Approval
	for rows.Next() {
		approval := models.Approval{}
		var approvedBy sql.NullString
		var approvedAt sql.NullTime
		err := rows.Scan(
			&approval.ID, &approval.ExpenseID, &approvedBy, &approval.Status,
			&approval.Comment, &approval.CreatedAt, &approvedAt,
		)
		if err != nil {
			return nil, err
		}
		if approvedBy.Valid {
			uid, _ := uuid.Parse(approvedBy.String)
			approval.ApprovedBy = uid
		}
		if approvedAt.Valid {
			approval.ApprovedAt = &approvedAt.Time
		}
		approvals = append(approvals, approval)
	}
	return approvals, nil
}

func (r *ApprovalRepository) UpdateStatus(id uuid.UUID, status models.ApprovalStatus, approvedBy uuid.UUID, comment string) error {
	now := time.Now()
	query := `
		UPDATE approvals SET status = $1, approved_by = $2, comment = $3, approved_at = $4
		WHERE id = $5
	`
	result, err := r.db.Exec(query, status, approvedBy, comment, now, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrApprovalNotFound
	}
	return nil
}
