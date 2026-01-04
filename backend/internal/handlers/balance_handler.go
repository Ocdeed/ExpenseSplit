package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/expensesplit/backend/internal/models"
	"github.com/expensesplit/backend/internal/services"
	"github.com/expensesplit/backend/pkg/utils"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type BalanceHandler struct {
	balanceService *services.BalanceService
	teamService    *services.TeamService
}

func NewBalanceHandler(balanceService *services.BalanceService, teamService *services.TeamService) *BalanceHandler {
	return &BalanceHandler{
		balanceService: balanceService,
		teamService:    teamService,
	}
}

func (h *BalanceHandler) GetTeamBalances(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		utils.Unauthorized(w, "User not authenticated")
		return
	}

	vars := mux.Vars(r)
	teamID, err := uuid.Parse(vars["teamId"])
	if err != nil {
		utils.BadRequest(w, "Invalid team ID")
		return
	}

	// Check if user is a member
	isMember, err := h.teamService.IsMember(teamID, userID)
	if err != nil {
		utils.InternalError(w, "Failed to check membership")
		return
	}
	if !isMember {
		utils.Forbidden(w, "You are not a member of this team")
		return
	}

	balances, err := h.balanceService.CalculateBalances(teamID)
	if err != nil {
		utils.InternalError(w, "Failed to calculate balances")
		return
	}

	utils.Success(w, balances, "")
}

func (h *BalanceHandler) GetUserBalance(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		utils.Unauthorized(w, "User not authenticated")
		return
	}

	vars := mux.Vars(r)
	teamID, err := uuid.Parse(vars["teamId"])
	if err != nil {
		utils.BadRequest(w, "Invalid team ID")
		return
	}

	// Check if user is a member
	isMember, err := h.teamService.IsMember(teamID, userID)
	if err != nil {
		utils.InternalError(w, "Failed to check membership")
		return
	}
	if !isMember {
		utils.Forbidden(w, "You are not a member of this team")
		return
	}

	balance, err := h.balanceService.GetUserBalance(teamID, userID)
	if err != nil {
		utils.InternalError(w, "Failed to get balance")
		return
	}

	utils.Success(w, balance, "")
}

func (h *BalanceHandler) RecordSettlement(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		utils.Unauthorized(w, "User not authenticated")
		return
	}

	vars := mux.Vars(r)
	teamID, err := uuid.Parse(vars["teamId"])
	if err != nil {
		utils.BadRequest(w, "Invalid team ID")
		return
	}

	// Check if user is a member
	isMember, err := h.teamService.IsMember(teamID, userID)
	if err != nil {
		utils.InternalError(w, "Failed to check membership")
		return
	}
	if !isMember {
		utils.Forbidden(w, "You are not a member of this team")
		return
	}

	var req models.SettlementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	// Validate that the user is either the payer or receiver
	if req.FromUser != userID && req.ToUser != userID {
		utils.Forbidden(w, "You can only record settlements involving yourself")
		return
	}

	if req.Amount <= 0 {
		utils.BadRequest(w, "Amount must be greater than 0")
		return
	}

	err = h.balanceService.RecordSettlement(teamID, &req)
	if err != nil {
		utils.InternalError(w, "Failed to record settlement")
		return
	}

	// Return updated balances
	balances, err := h.balanceService.CalculateBalances(teamID)
	if err != nil {
		utils.InternalError(w, "Failed to get updated balances")
		return
	}

	utils.Success(w, balances, "Settlement recorded successfully")
}
