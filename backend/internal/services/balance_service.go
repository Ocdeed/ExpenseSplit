package services

import (
	"github.com/expensesplit/backend/internal/models"
	"github.com/expensesplit/backend/internal/repository"
	"github.com/google/uuid"
)

type BalanceService struct {
	expenseRepo    *repository.ExpenseRepository
	teamRepo       *repository.TeamRepository
	userRepo       *repository.UserRepository
	settlementRepo *repository.SettlementRepository
}

func NewBalanceService(
	expenseRepo *repository.ExpenseRepository,
	teamRepo *repository.TeamRepository,
	userRepo *repository.UserRepository,
	settlementRepo *repository.SettlementRepository,
) *BalanceService {
	return &BalanceService{
		expenseRepo:    expenseRepo,
		teamRepo:       teamRepo,
		userRepo:       userRepo,
		settlementRepo: settlementRepo,
	}
}

// CalculateBalances calculates who owes whom in a team
func (s *BalanceService) CalculateBalances(teamID uuid.UUID) (*models.TeamBalanceSummary, error) {
	team, err := s.teamRepo.GetByID(teamID)
	if err != nil {
		return nil, err
	}

	members, err := s.teamRepo.GetTeamMembers(teamID)
	if err != nil {
		return nil, err
	}

	// Get all expenses for the team
	expenses, _, err := s.expenseRepo.GetByTeamID(teamID, 10000, 0) // Get all expenses
	if err != nil {
		return nil, err
	}

	// Get all settlements
	settlements, err := s.settlementRepo.GetByTeamID(teamID)
	if err != nil {
		return nil, err
	}

	// Calculate net balances between users
	// balanceMap[fromUser][toUser] = amount (positive means fromUser owes toUser)
	balanceMap := make(map[uuid.UUID]map[uuid.UUID]float64)

	// Initialize balance map for all members
	for _, member := range members {
		balanceMap[member.UserID] = make(map[uuid.UUID]float64)
	}

	// Process expenses
	for _, expense := range expenses {
		splits, err := s.expenseRepo.GetSplitsByExpenseID(expense.ID)
		if err != nil {
			return nil, err
		}

		for _, split := range splits {
			if split.UserID != expense.PaidBy && !split.IsSettled {
				// This user owes the payer
				if balanceMap[split.UserID] == nil {
					balanceMap[split.UserID] = make(map[uuid.UUID]float64)
				}
				balanceMap[split.UserID][expense.PaidBy] += split.Amount
			}
		}
	}

	// Process settlements (reduce balances)
	for _, settlement := range settlements {
		if balanceMap[settlement.FromUser] != nil {
			balanceMap[settlement.FromUser][settlement.ToUser] -= settlement.Amount
		}
	}

	// Simplify balances (net out mutual debts)
	simplifiedBalances := s.simplifyBalances(balanceMap)

	// Build response
	var balances []models.BalanceResponse
	memberSummaries := make(map[uuid.UUID]*models.UserBalanceSummary)

	// Initialize member summaries
	for _, member := range members {
		user, err := s.userRepo.GetByID(member.UserID)
		if err != nil {
			continue
		}
		memberSummaries[member.UserID] = &models.UserBalanceSummary{
			User:       user.ToResponse(),
			TotalOwed:  0,
			TotalOwing: 0,
			NetBalance: 0,
		}
	}

	// Build balance responses and update summaries
	for fromUserID, toUsers := range simplifiedBalances {
		for toUserID, amount := range toUsers {
			if amount > 0.01 { // Only include non-zero balances
				fromUser, err := s.userRepo.GetByID(fromUserID)
				if err != nil {
					continue
				}
				toUser, err := s.userRepo.GetByID(toUserID)
				if err != nil {
					continue
				}

				balances = append(balances, models.BalanceResponse{
					FromUser: fromUser.ToResponse(),
					ToUser:   toUser.ToResponse(),
					Amount:   amount,
				})

				// Update summaries
				if summary, ok := memberSummaries[fromUserID]; ok {
					summary.TotalOwed += amount
					summary.NetBalance -= amount
				}
				if summary, ok := memberSummaries[toUserID]; ok {
					summary.TotalOwing += amount
					summary.NetBalance += amount
				}
			}
		}
	}

	// Convert map to slice
	var memberSummarySlice []models.UserBalanceSummary
	for _, summary := range memberSummaries {
		memberSummarySlice = append(memberSummarySlice, *summary)
	}

	return &models.TeamBalanceSummary{
		TeamID:   teamID,
		TeamName: team.Name,
		Balances: balances,
		Members:  memberSummarySlice,
	}, nil
}

// simplifyBalances nets out mutual debts
func (s *BalanceService) simplifyBalances(balanceMap map[uuid.UUID]map[uuid.UUID]float64) map[uuid.UUID]map[uuid.UUID]float64 {
	simplified := make(map[uuid.UUID]map[uuid.UUID]float64)

	for fromUser, toUsers := range balanceMap {
		for toUser, amount := range toUsers {
			if amount <= 0 {
				continue
			}

			// Check if there's a reverse debt
			reverseAmount := float64(0)
			if balanceMap[toUser] != nil {
				reverseAmount = balanceMap[toUser][fromUser]
			}

			netAmount := amount - reverseAmount
			if netAmount > 0 {
				if simplified[fromUser] == nil {
					simplified[fromUser] = make(map[uuid.UUID]float64)
				}
				simplified[fromUser][toUser] = netAmount
			} else if netAmount < 0 {
				if simplified[toUser] == nil {
					simplified[toUser] = make(map[uuid.UUID]float64)
				}
				simplified[toUser][fromUser] = -netAmount
			}
		}
	}

	return simplified
}

// RecordSettlement records a settlement between two users
func (s *BalanceService) RecordSettlement(teamID uuid.UUID, req *models.SettlementRequest) error {
	settlement := &models.Settlement{
		TeamID:   teamID,
		FromUser: req.FromUser,
		ToUser:   req.ToUser,
		Amount:   req.Amount,
	}

	return s.settlementRepo.Create(settlement)
}

// GetUserBalance gets the balance summary for a specific user in a team
func (s *BalanceService) GetUserBalance(teamID, userID uuid.UUID) (*models.UserBalanceSummary, error) {
	teamSummary, err := s.CalculateBalances(teamID)
	if err != nil {
		return nil, err
	}

	for _, member := range teamSummary.Members {
		if member.User.ID == userID {
			return &member, nil
		}
	}

	// User not found in team, return empty summary
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	return &models.UserBalanceSummary{
		User:       user.ToResponse(),
		TotalOwed:  0,
		TotalOwing: 0,
		NetBalance: 0,
	}, nil
}
