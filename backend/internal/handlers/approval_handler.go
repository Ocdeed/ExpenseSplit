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

type ApprovalHandler struct {
	approvalService *services.ApprovalService
	teamService     *services.TeamService
}

func NewApprovalHandler(approvalService *services.ApprovalService, teamService *services.TeamService) *ApprovalHandler {
	return &ApprovalHandler{
		approvalService: approvalService,
		teamService:     teamService,
	}
}

func (h *ApprovalHandler) GetTeamApprovals(w http.ResponseWriter, r *http.Request) {
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

	approvals, err := h.approvalService.GetTeamApprovals(teamID)
	if err != nil {
		utils.InternalError(w, "Failed to get approvals")
		return
	}

	utils.Success(w, approvals, "Approvals retrieved successfully")
}

func (h *ApprovalHandler) UpdateApprovalStatus(w http.ResponseWriter, r *http.Request) {
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
	approvalID, err := uuid.Parse(vars["id"])
	if err != nil {
		utils.BadRequest(w, "Invalid approval ID")
		return
	}

	// Check if user is an admin
	isAdmin, err := h.teamService.IsAdmin(teamID, userID)
	if err != nil {
		utils.InternalError(w, "Failed to check permissions")
		return
	}
	if !isAdmin {
		utils.Forbidden(w, "Only admins can update approval status")
		return
	}

	var req models.ApprovalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	if err := h.approvalService.UpdateApprovalStatus(approvalID, userID, req.Status, req.Comment); err != nil {
		utils.InternalError(w, "Failed to update approval status")
		return
	}

	utils.Success(w, nil, "Approval status updated successfully")
}
