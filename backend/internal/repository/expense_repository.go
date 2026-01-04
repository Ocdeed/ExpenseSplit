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
	ErrExpenseNotFound = errors.New("expense not found")
)

type ExpenseRepository struct {
	db *database.DB
}

func NewExpenseRepository(db *database.DB) *ExpenseRepository {
	return &ExpenseRepository{db: db}
}

func (r *ExpenseRepository) Create(expense *models.Expense, splits []models.ExpenseSplit) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	expense.ID = uuid.New()
	expense.CreatedAt = time.Now()
	expense.UpdatedAt = time.Now()

	// Insert expense
	query := `
		INSERT INTO expenses (id, team_id, paid_by, amount, description, category, receipt_url, split_type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err = tx.Exec(query, expense.ID, expense.TeamID, expense.PaidBy, expense.Amount, expense.Description,
		expense.Category, expense.ReceiptURL, expense.SplitType, expense.CreatedAt, expense.UpdatedAt)
	if err != nil {
		return err
	}

	// Insert splits
	splitQuery := `
		INSERT INTO expense_splits (id, expense_id, user_id, amount, percent, is_settled)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	for i := range splits {
		splits[i].ID = uuid.New()
		splits[i].ExpenseID = expense.ID
		_, err = tx.Exec(splitQuery, splits[i].ID, splits[i].ExpenseID, splits[i].UserID,
			splits[i].Amount, splits[i].Percent, splits[i].IsSettled)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *ExpenseRepository) GetByID(id uuid.UUID) (*models.Expense, error) {
	expense := &models.Expense{}
	query := `
		SELECT id, team_id, paid_by, amount, description, category, receipt_url, split_type, created_at, updated_at
		FROM expenses WHERE id = $1
	`
	err := r.db.QueryRow(query, id).Scan(
		&expense.ID, &expense.TeamID, &expense.PaidBy, &expense.Amount, &expense.Description,
		&expense.Category, &expense.ReceiptURL, &expense.SplitType, &expense.CreatedAt, &expense.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrExpenseNotFound
	}
	if err != nil {
		return nil, err
	}
	return expense, nil
}

func (r *ExpenseRepository) GetByTeamID(teamID uuid.UUID, limit, offset int) ([]*models.Expense, int64, error) {
	// Get total count
	var total int64
	err := r.db.QueryRow("SELECT COUNT(*) FROM expenses WHERE team_id = $1", teamID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, team_id, paid_by, amount, description, category, receipt_url, split_type, created_at, updated_at
		FROM expenses WHERE team_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(query, teamID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var expenses []*models.Expense
	for rows.Next() {
		expense := &models.Expense{}
		err := rows.Scan(
			&expense.ID, &expense.TeamID, &expense.PaidBy, &expense.Amount, &expense.Description,
			&expense.Category, &expense.ReceiptURL, &expense.SplitType, &expense.CreatedAt, &expense.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		expenses = append(expenses, expense)
	}
	return expenses, total, nil
}

func (r *ExpenseRepository) GetSplitsByExpenseID(expenseID uuid.UUID) ([]models.ExpenseSplit, error) {
	query := `
		SELECT id, expense_id, user_id, amount, percent, is_settled
		FROM expense_splits WHERE expense_id = $1
	`
	rows, err := r.db.Query(query, expenseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var splits []models.ExpenseSplit
	for rows.Next() {
		split := models.ExpenseSplit{}
		err := rows.Scan(&split.ID, &split.ExpenseID, &split.UserID, &split.Amount, &split.Percent, &split.IsSettled)
		if err != nil {
			return nil, err
		}
		splits = append(splits, split)
	}
	return splits, nil
}

func (r *ExpenseRepository) Update(expense *models.Expense) error {
	expense.UpdatedAt = time.Now()
	query := `
		UPDATE expenses SET amount = $1, description = $2, category = $3, receipt_url = $4, updated_at = $5
		WHERE id = $6
	`
	result, err := r.db.Exec(query, expense.Amount, expense.Description, expense.Category,
		expense.ReceiptURL, expense.UpdatedAt, expense.ID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrExpenseNotFound
	}
	return nil
}

func (r *ExpenseRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM expenses WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrExpenseNotFound
	}
	return nil
}

func (r *ExpenseRepository) MarkSplitAsSettled(splitID uuid.UUID) error {
	query := `UPDATE expense_splits SET is_settled = true WHERE id = $1`
	result, err := r.db.Exec(query, splitID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("split not found")
	}
	return nil
}

func (r *ExpenseRepository) GetUnsettledSplitsByUser(teamID, userID uuid.UUID) ([]models.ExpenseSplit, error) {
	query := `
		SELECT es.id, es.expense_id, es.user_id, es.amount, es.percent, es.is_settled
		FROM expense_splits es
		INNER JOIN expenses e ON es.expense_id = e.id
		WHERE e.team_id = $1 AND es.user_id = $2 AND es.is_settled = false AND e.paid_by != $2
	`
	rows, err := r.db.Query(query, teamID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var splits []models.ExpenseSplit
	for rows.Next() {
		split := models.ExpenseSplit{}
		err := rows.Scan(&split.ID, &split.ExpenseID, &split.UserID, &split.Amount, &split.Percent, &split.IsSettled)
		if err != nil {
			return nil, err
		}
		splits = append(splits, split)
	}
	return splits, nil
}

func (r *ExpenseRepository) GetExpensesByUserPaid(teamID, userID uuid.UUID) ([]*models.Expense, error) {
	query := `
		SELECT id, team_id, paid_by, amount, description, category, receipt_url, split_type, created_at, updated_at
		FROM expenses WHERE team_id = $1 AND paid_by = $2
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query, teamID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []*models.Expense
	for rows.Next() {
		expense := &models.Expense{}
		err := rows.Scan(
			&expense.ID, &expense.TeamID, &expense.PaidBy, &expense.Amount, &expense.Description,
			&expense.Category, &expense.ReceiptURL, &expense.SplitType, &expense.CreatedAt, &expense.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		expenses = append(expenses, expense)
	}
	return expenses, nil
}
