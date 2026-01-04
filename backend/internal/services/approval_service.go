package services

import (
	"github.com/expensesplit/backend/internal/models"
	"github.com/expensesplit/backend/internal/repository"
	"github.com/google/uuid"
)

type ApprovalService struct {
	approvalRepo *repository.ApprovalRepository
	expenseRepo  *repository.ExpenseRepository
	teamRepo     *repository.TeamRepository
}

func NewApprovalService(
	approvalRepo *repository.ApprovalRepository,
	expenseRepo *repository.ExpenseRepository,
	teamRepo *repository.TeamRepository,
) *ApprovalService {
	return &ApprovalService{
		approvalRepo: approvalRepo,
		expenseRepo:  expenseRepo,
		teamRepo:     teamRepo,
	}
}

func (s *ApprovalService) CreateApproval(expenseID uuid.UUID) (*models.Approval, error) {
	approval := &models.Approval{
		ExpenseID: expenseID,
		Status:    models.ApprovalStatusPending,
	}
	if err := s.approvalRepo.Create(approval); err != nil {
		return nil, err
	}
	return approval, nil
}

func (s *ApprovalService) UpdateApprovalStatus(approvalID, userID uuid.UUID, status models.ApprovalStatus, comment string) error {
	return s.approvalRepo.UpdateStatus(approvalID, status, userID, comment)
}

func (s *ApprovalService) GetTeamApprovals(teamID uuid.UUID) ([]models.Approval, error) {
	return s.approvalRepo.GetAllByTeamID(teamID)
}
