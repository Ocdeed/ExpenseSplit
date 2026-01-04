package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/expensesplit/backend/internal/models"
	"github.com/expensesplit/backend/internal/repository"
	"github.com/expensesplit/backend/internal/services"
	"github.com/expensesplit/backend/pkg/utils"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type TeamHandler struct {
	teamService *services.TeamService
}

func NewTeamHandler(teamService *services.TeamService) *TeamHandler {
	return &TeamHandler{teamService: teamService}
}

func (h *TeamHandler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		utils.Unauthorized(w, "User not authenticated")
		return
	}

	var req models.TeamCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	team, err := h.teamService.CreateTeam(&req, userID)
	if err != nil {
		if err == services.ErrTeamNameRequired {
			utils.BadRequest(w, err.Error())
			return
		}
		utils.InternalError(w, "Failed to create team")
		return
	}

	utils.Created(w, team, "Team created successfully")
}

func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		utils.Unauthorized(w, "User not authenticated")
		return
	}

	vars := mux.Vars(r)
	teamID, err := uuid.Parse(vars["id"])
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

	team, err := h.teamService.GetTeamWithMembers(teamID)
	if err != nil {
		if err == repository.ErrTeamNotFound {
			utils.NotFound(w, "Team not found")
			return
		}
		utils.InternalError(w, "Failed to get team")
		return
	}

	utils.Success(w, team, "")
}

func (h *TeamHandler) GetUserTeams(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		utils.Unauthorized(w, "User not authenticated")
		return
	}

	teams, err := h.teamService.GetUserTeams(userID)
	if err != nil {
		utils.InternalError(w, "Failed to get teams")
		return
	}

	if teams == nil {
		teams = []*models.TeamResponse{}
	}

	utils.Success(w, teams, "")
}

func (h *TeamHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		utils.Unauthorized(w, "User not authenticated")
		return
	}

	vars := mux.Vars(r)
	teamID, err := uuid.Parse(vars["id"])
	if err != nil {
		utils.BadRequest(w, "Invalid team ID")
		return
	}

	var req models.AddMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	if req.Email == "" {
		utils.BadRequest(w, "Email is required")
		return
	}

	err = h.teamService.AddMember(teamID, &req, userID)
	if err != nil {
		switch err {
		case services.ErrNotAuthorized:
			utils.Forbidden(w, "Only admins can add members")
		case repository.ErrUserNotFound:
			utils.NotFound(w, "User not found")
		case repository.ErrAlreadyMember:
			utils.BadRequest(w, "User is already a member")
		default:
			utils.InternalError(w, "Failed to add member")
		}
		return
	}

	team, _ := h.teamService.GetTeamWithMembers(teamID)
	utils.Success(w, team, "Member added successfully")
}

func (h *TeamHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		utils.Unauthorized(w, "User not authenticated")
		return
	}

	vars := mux.Vars(r)
	teamID, err := uuid.Parse(vars["id"])
	if err != nil {
		utils.BadRequest(w, "Invalid team ID")
		return
	}

	memberID, err := uuid.Parse(vars["memberId"])
	if err != nil {
		utils.BadRequest(w, "Invalid member ID")
		return
	}

	err = h.teamService.RemoveMember(teamID, memberID, userID)
	if err != nil {
		switch err {
		case services.ErrNotAuthorized:
			utils.Forbidden(w, "Only admins can remove members")
		case repository.ErrNotTeamMember:
			utils.NotFound(w, "Member not found")
		case repository.ErrCannotRemoveOwner:
			utils.BadRequest(w, "Cannot remove team owner")
		default:
			utils.InternalError(w, "Failed to remove member")
		}
		return
	}

	utils.Success(w, nil, "Member removed successfully")
}

func (h *TeamHandler) UpdateTeam(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		utils.Unauthorized(w, "User not authenticated")
		return
	}

	vars := mux.Vars(r)
	teamID, err := uuid.Parse(vars["id"])
	if err != nil {
		utils.BadRequest(w, "Invalid team ID")
		return
	}

	var req models.TeamCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	team, err := h.teamService.UpdateTeam(teamID, req.Name, userID)
	if err != nil {
		switch err {
		case services.ErrNotAuthorized:
			utils.Forbidden(w, "Only admins can update team")
		case repository.ErrTeamNotFound:
			utils.NotFound(w, "Team not found")
		default:
			utils.InternalError(w, "Failed to update team")
		}
		return
	}

	utils.Success(w, team, "Team updated successfully")
}

func (h *TeamHandler) DeleteTeam(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		utils.Unauthorized(w, "User not authenticated")
		return
	}

	vars := mux.Vars(r)
	teamID, err := uuid.Parse(vars["id"])
	if err != nil {
		utils.BadRequest(w, "Invalid team ID")
		return
	}

	err = h.teamService.DeleteTeam(teamID, userID)
	if err != nil {
		switch err {
		case services.ErrNotAuthorized:
			utils.Forbidden(w, "Only team owner can delete team")
		case repository.ErrTeamNotFound:
			utils.NotFound(w, "Team not found")
		default:
			utils.InternalError(w, "Failed to delete team")
		}
		return
	}

	utils.Success(w, nil, "Team deleted successfully")
}

func (h *TeamHandler) GetTeamMembers(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		utils.Unauthorized(w, "User not authenticated")
		return
	}

	vars := mux.Vars(r)
	teamID, err := uuid.Parse(vars["id"])
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

	members, err := h.teamService.GetTeamMembers(teamID)
	if err != nil {
		utils.InternalError(w, "Failed to get members")
		return
	}

	utils.Success(w, members, "")
}
