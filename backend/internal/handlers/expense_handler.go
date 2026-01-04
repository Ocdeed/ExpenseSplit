package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/expensesplit/backend/internal/models"
	"github.com/expensesplit/backend/internal/repository"
	"github.com/expensesplit/backend/internal/services"
	"github.com/expensesplit/backend/pkg/utils"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type ExpenseHandler struct {
	expenseService *services.ExpenseService
	teamService    *services.TeamService
	uploadDir      string
}

func NewExpenseHandler(expenseService *services.ExpenseService, teamService *services.TeamService, uploadDir string) *ExpenseHandler {
	return &ExpenseHandler{
		expenseService: expenseService,
		teamService:    teamService,
		uploadDir:      uploadDir,
	}
}

func (h *ExpenseHandler) CreateExpense(w http.ResponseWriter, r *http.Request) {
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

	var req models.ExpenseCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	expense, err := h.expenseService.CreateExpense(teamID, userID, &req)
	if err != nil {
		switch err {
		case services.ErrAmountRequired, services.ErrSplitWithRequired, services.ErrInvalidSplitType, services.ErrInvalidCustomSplit:
			utils.BadRequest(w, err.Error())
		default:
			utils.InternalError(w, "Failed to create expense")
		}
		return
	}

	utils.Created(w, expense, "Expense created successfully")
}

func (h *ExpenseHandler) GetExpense(w http.ResponseWriter, r *http.Request) {
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

	expenseID, err := uuid.Parse(vars["id"])
	if err != nil {
		utils.BadRequest(w, "Invalid expense ID")
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

	expense, err := h.expenseService.GetExpenseByID(expenseID)
	if err != nil {
		if err == repository.ErrExpenseNotFound {
			utils.NotFound(w, "Expense not found")
			return
		}
		utils.InternalError(w, "Failed to get expense")
		return
	}

	utils.Success(w, expense, "")
}

func (h *ExpenseHandler) GetTeamExpenses(w http.ResponseWriter, r *http.Request) {
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

	// Parse pagination params
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}

	expenses, total, err := h.expenseService.GetTeamExpenses(teamID, page, perPage)
	if err != nil {
		utils.InternalError(w, "Failed to get expenses")
		return
	}

	if expenses == nil {
		expenses = []*models.ExpenseResponse{}
	}

	utils.Paginated(w, expenses, page, perPage, total)
}

func (h *ExpenseHandler) UpdateExpense(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		utils.Unauthorized(w, "User not authenticated")
		return
	}

	vars := mux.Vars(r)
	expenseID, err := uuid.Parse(vars["id"])
	if err != nil {
		utils.BadRequest(w, "Invalid expense ID")
		return
	}

	var req models.ExpenseUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	expense, err := h.expenseService.UpdateExpense(expenseID, &req, userID)
	if err != nil {
		switch err {
		case services.ErrNotAuthorized:
			utils.Forbidden(w, "Only the payer can update this expense")
		case repository.ErrExpenseNotFound:
			utils.NotFound(w, "Expense not found")
		default:
			utils.InternalError(w, "Failed to update expense")
		}
		return
	}

	utils.Success(w, expense, "Expense updated successfully")
}

func (h *ExpenseHandler) DeleteExpense(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		utils.Unauthorized(w, "User not authenticated")
		return
	}

	vars := mux.Vars(r)
	expenseID, err := uuid.Parse(vars["id"])
	if err != nil {
		utils.BadRequest(w, "Invalid expense ID")
		return
	}

	err = h.expenseService.DeleteExpense(expenseID, userID)
	if err != nil {
		switch err {
		case services.ErrNotAuthorized:
			utils.Forbidden(w, "Only the payer can delete this expense")
		case repository.ErrExpenseNotFound:
			utils.NotFound(w, "Expense not found")
		default:
			utils.InternalError(w, "Failed to delete expense")
		}
		return
	}

	utils.Success(w, nil, "Expense deleted successfully")
}

func (h *ExpenseHandler) UploadReceipt(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		utils.Unauthorized(w, "User not authenticated")
		return
	}

	vars := mux.Vars(r)
	expenseID, err := uuid.Parse(vars["id"])
	if err != nil {
		utils.BadRequest(w, "Invalid expense ID")
		return
	}

	// Parse multipart form (max 10MB)
	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		utils.BadRequest(w, "Failed to parse form data")
		return
	}

	file, header, err := r.FormFile("receipt")
	if err != nil {
		utils.BadRequest(w, "Receipt file is required")
		return
	}
	defer file.Close()

	// Validate file type
	allowedTypes := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".pdf":  true,
	}
	ext := filepath.Ext(header.Filename)
	if !allowedTypes[ext] {
		utils.BadRequest(w, "Invalid file type. Allowed: jpg, jpeg, png, pdf")
		return
	}

	// Create upload directory if it doesn't exist
	if err := os.MkdirAll(h.uploadDir, 0755); err != nil {
		utils.InternalError(w, "Failed to create upload directory")
		return
	}

	// Generate unique filename
	filename := uuid.New().String() + ext
	filePath := filepath.Join(h.uploadDir, filename)

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		utils.InternalError(w, "Failed to save file")
		return
	}
	defer dst.Close()

	// Copy file content
	if _, err := io.Copy(dst, file); err != nil {
		utils.InternalError(w, "Failed to save file")
		return
	}

	// Update expense with receipt URL
	receiptURL := "/uploads/" + filename
	if err := h.expenseService.UpdateReceiptURL(expenseID, receiptURL); err != nil {
		// Clean up file on error
		os.Remove(filePath)
		if err == repository.ErrExpenseNotFound {
			utils.NotFound(w, "Expense not found")
			return
		}
		utils.InternalError(w, "Failed to update expense")
		return
	}

	// Ignore userID for now (used for authorization in future)
	_ = userID

	utils.Success(w, map[string]string{"receipt_url": receiptURL}, "Receipt uploaded successfully")
}
