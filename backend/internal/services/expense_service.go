package services

import (
	"errors"

	"github.com/expensesplit/backend/internal/models"
	"github.com/expensesplit/backend/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrAmountRequired     = errors.New("amount is required and must be greater than 0")
	ErrSplitWithRequired  = errors.New("at least one user to split with is required")
	ErrInvalidSplitType   = errors.New("invalid split type")
	ErrInvalidCustomSplit = errors.New("custom split amounts must equal total amount")
)

type ExpenseService struct {
	expenseRepo  *repository.ExpenseRepository
	teamRepo     *repository.TeamRepository
	userRepo     *repository.UserRepository
	approvalRepo *repository.ApprovalRepository
}

func NewExpenseService(
	expenseRepo *repository.ExpenseRepository,
	teamRepo *repository.TeamRepository,
	userRepo *repository.UserRepository,
	approvalRepo *repository.ApprovalRepository,
) *ExpenseService {
	return &ExpenseService{
		expenseRepo:  expenseRepo,
		teamRepo:     teamRepo,
		userRepo:     userRepo,
		approvalRepo: approvalRepo,
	}
}

func (s *ExpenseService) CreateExpense(teamID, paidBy uuid.UUID, req *models.ExpenseCreateRequest) (*models.ExpenseResponse, error) {
	// Validate input
	if req.Amount <= 0 {
		return nil, ErrAmountRequired
	}
	if len(req.SplitWith) == 0 {
		return nil, ErrSplitWithRequired
	}

	// Validate split type
	if req.SplitType == "" {
		req.SplitType = models.SplitTypeEqual
	}
	if req.SplitType != models.SplitTypeEqual && req.SplitType != models.SplitTypeCustom && req.SplitType != models.SplitTypePercent {
		return nil, ErrInvalidSplitType
	}

	// Create expense
	expense := &models.Expense{
		TeamID:      teamID,
		PaidBy:      paidBy,
		Amount:      req.Amount,
		Description: req.Description,
		Category:    req.Category,
		SplitType:   req.SplitType,
	}

	// Calculate splits
	splits, err := s.calculateSplits(expense, req)
	if err != nil {
		return nil, err
	}

	// Save expense and splits
	if err := s.expenseRepo.Create(expense, splits); err != nil {
		return nil, err
	}

	// Create approval record
	approval := &models.Approval{
		ExpenseID: expense.ID,
		Status:    models.ApprovalStatusPending,
	}
	if err := s.approvalRepo.Create(approval); err != nil {
		// Log error but don't fail expense creation
		// In a real app, you might want to use a transaction
	}

	return s.GetExpenseByID(expense.ID)
}

func (s *ExpenseService) calculateSplits(expense *models.Expense, req *models.ExpenseCreateRequest) ([]models.ExpenseSplit, error) {
	var splits []models.ExpenseSplit

	switch req.SplitType {
	case models.SplitTypeEqual:
		// Equal split among all users (including the payer)
		numUsers := len(req.SplitWith)
		splitAmount := expense.Amount / float64(numUsers)
		splitPercent := 100.0 / float64(numUsers)

		for _, userID := range req.SplitWith {
			splits = append(splits, models.ExpenseSplit{
				UserID:    userID,
				Amount:    splitAmount,
				Percent:   splitPercent,
				IsSettled: userID == expense.PaidBy, // Payer's share is already settled
			})
		}

	case models.SplitTypeCustom:
		// Custom amount split
		var totalCustom float64
		for _, entry := range req.CustomSplit {
			totalCustom += entry.Amount
		}
		if totalCustom != expense.Amount {
			return nil, ErrInvalidCustomSplit
		}

		for _, entry := range req.CustomSplit {
			percent := (entry.Amount / expense.Amount) * 100
			splits = append(splits, models.ExpenseSplit{
				UserID:    entry.UserID,
				Amount:    entry.Amount,
				Percent:   percent,
				IsSettled: entry.UserID == expense.PaidBy,
			})
		}

	case models.SplitTypePercent:
		// Percentage split
		var totalPercent float64
		for _, entry := range req.CustomSplit {
			totalPercent += entry.Percent
		}
		if totalPercent != 100 {
			return nil, errors.New("percentages must add up to 100")
		}

		for _, entry := range req.CustomSplit {
			amount := (entry.Percent / 100) * expense.Amount
			splits = append(splits, models.ExpenseSplit{
				UserID:    entry.UserID,
				Amount:    amount,
				Percent:   entry.Percent,
				IsSettled: entry.UserID == expense.PaidBy,
			})
		}
	}

	return splits, nil
}

func (s *ExpenseService) GetExpenseByID(id uuid.UUID) (*models.ExpenseResponse, error) {
	expense, err := s.expenseRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return s.buildExpenseResponse(expense)
}

func (s *ExpenseService) buildExpenseResponse(expense *models.Expense) (*models.ExpenseResponse, error) {
	// Get payer info
	payer, err := s.userRepo.GetByID(expense.PaidBy)
	if err != nil {
		return nil, err
	}

	// Get splits
	splits, err := s.expenseRepo.GetSplitsByExpenseID(expense.ID)
	if err != nil {
		return nil, err
	}

	// Build split details
	var splitDetails []models.ExpenseSplitDetail
	for _, split := range splits {
		user, err := s.userRepo.GetByID(split.UserID)
		if err != nil {
			return nil, err
		}
		splitDetails = append(splitDetails, models.ExpenseSplitDetail{
			ID:        split.ID,
			User:      user.ToResponse(),
			Amount:    split.Amount,
			Percent:   split.Percent,
			IsSettled: split.IsSettled,
		})
	}

	// Get approval status
	approval, err := s.approvalRepo.GetByExpenseID(expense.ID)
	status := models.ApprovalStatusPending
	if err == nil && approval != nil {
		status = approval.Status
	}

	return &models.ExpenseResponse{
		ID:             expense.ID,
		TeamID:         expense.TeamID,
		PaidBy:         payer.ToResponse(),
		Amount:         expense.Amount,
		Description:    expense.Description,
		Category:       expense.Category,
		ReceiptURL:     expense.ReceiptURL,
		SplitType:      expense.SplitType,
		Splits:         splitDetails,
		ApprovalStatus: status,
		CreatedAt:      expense.CreatedAt,
	}, nil
}

func (s *ExpenseService) GetTeamExpenses(teamID uuid.UUID, page, perPage int) ([]*models.ExpenseResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	expenses, total, err := s.expenseRepo.GetByTeamID(teamID, perPage, offset)
	if err != nil {
		return nil, 0, err
	}

	var responses []*models.ExpenseResponse
	for _, expense := range expenses {
		response, err := s.buildExpenseResponse(expense)
		if err != nil {
			return nil, 0, err
		}
		responses = append(responses, response)
	}

	return responses, total, nil
}

func (s *ExpenseService) UpdateExpense(id uuid.UUID, req *models.ExpenseUpdateRequest, requesterID uuid.UUID) (*models.ExpenseResponse, error) {
	expense, err := s.expenseRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Only the payer can update the expense
	if expense.PaidBy != requesterID {
		return nil, ErrNotAuthorized
	}

	if req.Amount != nil {
		expense.Amount = *req.Amount
	}
	if req.Description != nil {
		expense.Description = *req.Description
	}
	if req.Category != nil {
		expense.Category = *req.Category
	}

	if err := s.expenseRepo.Update(expense); err != nil {
		return nil, err
	}

	return s.GetExpenseByID(id)
}

func (s *ExpenseService) DeleteExpense(id uuid.UUID, requesterID uuid.UUID) error {
	expense, err := s.expenseRepo.GetByID(id)
	if err != nil {
		return err
	}

	// Only the payer can delete the expense
	if expense.PaidBy != requesterID {
		return ErrNotAuthorized
	}

	return s.expenseRepo.Delete(id)
}

func (s *ExpenseService) UpdateReceiptURL(id uuid.UUID, receiptURL string) error {
	expense, err := s.expenseRepo.GetByID(id)
	if err != nil {
		return err
	}
	expense.ReceiptURL = receiptURL
	return s.expenseRepo.Update(expense)
}

func (s *ExpenseService) MarkSplitAsSettled(splitID uuid.UUID) error {
	return s.expenseRepo.MarkSplitAsSettled(splitID)
}
