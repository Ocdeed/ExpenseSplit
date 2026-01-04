package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/expensesplit/backend/internal/models"
	"github.com/expensesplit/backend/internal/repository"
	"github.com/expensesplit/backend/internal/services"
	"github.com/expensesplit/backend/pkg/utils"
)

type AuthHandler struct {
	authService *services.AuthService
	userRepo    *repository.UserRepository
}

func NewAuthHandler(authService *services.AuthService, userRepo *repository.UserRepository) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userRepo:    userRepo,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.UserCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	response, err := h.authService.Register(&req)
	if err != nil {
		switch err {
		case services.ErrEmailRequired, services.ErrPasswordRequired, services.ErrNameRequired, services.ErrPasswordTooShort:
			utils.BadRequest(w, err.Error())
		case repository.ErrUserAlreadyExists:
			utils.BadRequest(w, "User with this email already exists")
		default:
			utils.InternalError(w, "Failed to register user")
		}
		return
	}

	utils.Created(w, response, "User registered successfully")
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.UserLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	response, err := h.authService.Login(&req)
	if err != nil {
		switch err {
		case services.ErrEmailRequired, services.ErrPasswordRequired:
			utils.BadRequest(w, err.Error())
		case services.ErrInvalidCredentials:
			utils.Unauthorized(w, "Invalid email or password")
		default:
			utils.InternalError(w, "Failed to login")
		}
		return
	}

	utils.Success(w, response, "Login successful")
}

func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		utils.Unauthorized(w, "User not authenticated")
		return
	}

	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		utils.InternalError(w, "Failed to get user")
		return
	}

	utils.Success(w, user.ToResponse(), "")
}
