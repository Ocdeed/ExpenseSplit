package services

import (
	"errors"

	"github.com/expensesplit/backend/internal/models"
	"github.com/expensesplit/backend/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrTeamNameRequired = errors.New("team name is required")
	ErrNotAuthorized    = errors.New("not authorized to perform this action")
)

type TeamService struct {
	teamRepo *repository.TeamRepository
	userRepo *repository.UserRepository
}

func NewTeamService(teamRepo *repository.TeamRepository, userRepo *repository.UserRepository) *TeamService {
	return &TeamService{
		teamRepo: teamRepo,
		userRepo: userRepo,
	}
}

func (s *TeamService) CreateTeam(req *models.TeamCreateRequest, creatorID uuid.UUID) (*models.TeamResponse, error) {
	if req.Name == "" {
		return nil, ErrTeamNameRequired
	}

	team := &models.Team{
		Name: req.Name,
	}

	if err := s.teamRepo.Create(team, creatorID); err != nil {
		return nil, err
	}

	// Get team with members
	return s.GetTeamWithMembers(team.ID)
}

func (s *TeamService) GetTeamWithMembers(teamID uuid.UUID) (*models.TeamResponse, error) {
	team, err := s.teamRepo.GetByID(teamID)
	if err != nil {
		return nil, err
	}

	members, err := s.teamRepo.GetTeamMembers(teamID)
	if err != nil {
		return nil, err
	}

	response := team.ToResponse()
	response.Members = members
	return &response, nil
}

func (s *TeamService) GetUserTeams(userID uuid.UUID) ([]*models.TeamResponse, error) {
	teams, err := s.teamRepo.GetUserTeams(userID)
	if err != nil {
		return nil, err
	}

	var responses []*models.TeamResponse
	for _, team := range teams {
		members, err := s.teamRepo.GetTeamMembers(team.ID)
		if err != nil {
			return nil, err
		}
		response := team.ToResponse()
		response.Members = members
		responses = append(responses, &response)
	}
	return responses, nil
}

func (s *TeamService) AddMember(teamID uuid.UUID, req *models.AddMemberRequest, requesterID uuid.UUID) error {
	// Check if requester is admin
	isAdmin, err := s.teamRepo.IsAdmin(teamID, requesterID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return ErrNotAuthorized
	}

	// Find user by email
	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		return err
	}

	role := req.Role
	if role == "" {
		role = "member"
	}

	return s.teamRepo.AddMember(teamID, user.ID, role)
}

func (s *TeamService) RemoveMember(teamID, userID, requesterID uuid.UUID) error {
	// Check if requester is admin
	isAdmin, err := s.teamRepo.IsAdmin(teamID, requesterID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return ErrNotAuthorized
	}

	return s.teamRepo.RemoveMember(teamID, userID)
}

func (s *TeamService) UpdateTeam(teamID uuid.UUID, name string, requesterID uuid.UUID) (*models.TeamResponse, error) {
	// Check if requester is admin
	isAdmin, err := s.teamRepo.IsAdmin(teamID, requesterID)
	if err != nil {
		return nil, err
	}
	if !isAdmin {
		return nil, ErrNotAuthorized
	}

	team, err := s.teamRepo.GetByID(teamID)
	if err != nil {
		return nil, err
	}

	team.Name = name
	if err := s.teamRepo.Update(team); err != nil {
		return nil, err
	}

	return s.GetTeamWithMembers(teamID)
}

func (s *TeamService) DeleteTeam(teamID, requesterID uuid.UUID) error {
	// Check if requester is the creator
	team, err := s.teamRepo.GetByID(teamID)
	if err != nil {
		return err
	}
	if team.CreatedBy != requesterID {
		return ErrNotAuthorized
	}

	return s.teamRepo.Delete(teamID)
}

func (s *TeamService) IsMember(teamID, userID uuid.UUID) (bool, error) {
	return s.teamRepo.IsMember(teamID, userID)
}

func (s *TeamService) IsAdmin(teamID, userID uuid.UUID) (bool, error) {
	return s.teamRepo.IsAdmin(teamID, userID)
}

func (s *TeamService) GetTeamMembers(teamID uuid.UUID) ([]models.MemberDetail, error) {
	return s.teamRepo.GetTeamMembers(teamID)
}
