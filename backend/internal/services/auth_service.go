package services

import (
	"errors"
	"time"

	"github.com/expensesplit/backend/internal/models"
	"github.com/expensesplit/backend/internal/repository"
	"github.com/expensesplit/backend/pkg/utils"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailRequired      = errors.New("email is required")
	ErrPasswordRequired   = errors.New("password is required")
	ErrNameRequired       = errors.New("name is required")
	ErrPasswordTooShort   = errors.New("password must be at least 6 characters")
)

type AuthService struct {
	userRepo   *repository.UserRepository
	jwtManager *utils.JWTManager
}

func NewAuthService(userRepo *repository.UserRepository, jwtSecret string, tokenDuration time.Duration) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtManager: utils.NewJWTManager(jwtSecret, tokenDuration),
	}
}

func (s *AuthService) Register(req *models.UserCreateRequest) (*models.AuthResponse, error) {
	// Validate input
	if req.Email == "" {
		return nil, ErrEmailRequired
	}
	if req.Password == "" {
		return nil, ErrPasswordRequired
	}
	if req.Name == "" {
		return nil, ErrNameRequired
	}
	if len(req.Password) < 6 {
		return nil, ErrPasswordTooShort
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &models.User{
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Name:         req.Name,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	// Generate token
	token, err := s.jwtManager.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		User:  user.ToResponse(),
		Token: token,
	}, nil
}

func (s *AuthService) Login(req *models.UserLoginRequest) (*models.AuthResponse, error) {
	// Validate input
	if req.Email == "" {
		return nil, ErrEmailRequired
	}
	if req.Password == "" {
		return nil, ErrPasswordRequired
	}

	// Get user by email
	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Check password
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	// Generate token
	token, err := s.jwtManager.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		User:  user.ToResponse(),
		Token: token,
	}, nil
}

func (s *AuthService) ValidateToken(tokenString string) (*utils.Claims, error) {
	return s.jwtManager.ValidateToken(tokenString)
}

func (s *AuthService) GetUserByID(id string) (*models.User, error) {
	return s.userRepo.GetByEmail(id)
}
