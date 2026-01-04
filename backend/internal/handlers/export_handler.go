package handlers

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"net/http"
	"time"

	"github.com/expensesplit/backend/internal/services"
	"github.com/expensesplit/backend/pkg/utils"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type ExportHandler struct {
	expenseService *services.ExpenseService
	balanceService *services.BalanceService
	teamService    *services.TeamService
}

func NewExportHandler(
	expenseService *services.ExpenseService,
	balanceService *services.BalanceService,
	teamService *services.TeamService,
) *ExportHandler {
	return &ExportHandler{
		expenseService: expenseService,
		balanceService: balanceService,
		teamService:    teamService,
	}
}

func (h *ExportHandler) ExportExpensesCSV(w http.ResponseWriter, r *http.Request) {
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

	// Get all expenses
	expenses, _, err := h.expenseService.GetTeamExpenses(teamID, 1, 10000)
	if err != nil {
		utils.InternalError(w, "Failed to get expenses")
		return
	}

	// Create CSV
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header
	header := []string{"ID", "Date", "Description", "Category", "Amount", "Paid By", "Split Type"}
	writer.Write(header)

	// Write data
	for _, expense := range expenses {
		row := []string{
			expense.ID.String(),
			expense.CreatedAt.Format("2006-01-02"),
			expense.Description,
			expense.Category,
			fmt.Sprintf("%.2f", expense.Amount),
			expense.PaidBy.Name,
			string(expense.SplitType),
		}
		writer.Write(row)
	}

	writer.Flush()

	// Set headers for file download
	filename := fmt.Sprintf("expenses_%s_%s.csv", teamID.String()[:8], time.Now().Format("20060102"))
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Write(buf.Bytes())
}

func (h *ExportHandler) ExportBalancesCSV(w http.ResponseWriter, r *http.Request) {
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

	// Get balances
	balances, err := h.balanceService.CalculateBalances(teamID)
	if err != nil {
		utils.InternalError(w, "Failed to calculate balances")
		return
	}

	// Create CSV
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header for balances
	writer.Write([]string{"From", "To", "Amount"})
	for _, balance := range balances.Balances {
		row := []string{
			balance.FromUser.Name,
			balance.ToUser.Name,
			fmt.Sprintf("%.2f", balance.Amount),
		}
		writer.Write(row)
	}

	// Add empty row
	writer.Write([]string{})

	// Write member summary
	writer.Write([]string{"Member Summary"})
	writer.Write([]string{"Name", "Total Owed", "Total Owing", "Net Balance"})
	for _, member := range balances.Members {
		row := []string{
			member.User.Name,
			fmt.Sprintf("%.2f", member.TotalOwed),
			fmt.Sprintf("%.2f", member.TotalOwing),
			fmt.Sprintf("%.2f", member.NetBalance),
		}
		writer.Write(row)
	}

	writer.Flush()

	// Set headers for file download
	filename := fmt.Sprintf("balances_%s_%s.csv", teamID.String()[:8], time.Now().Format("20060102"))
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Write(buf.Bytes())
}

func (h *ExportHandler) ExportReimbursementSummary(w http.ResponseWriter, r *http.Request) {
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

	// Get team info
	team, err := h.teamService.GetTeamWithMembers(teamID)
	if err != nil {
		utils.InternalError(w, "Failed to get team")
		return
	}

	// Get expenses
	expenses, _, err := h.expenseService.GetTeamExpenses(teamID, 1, 10000)
	if err != nil {
		utils.InternalError(w, "Failed to get expenses")
		return
	}

	// Get balances
	balances, err := h.balanceService.CalculateBalances(teamID)
	if err != nil {
		utils.InternalError(w, "Failed to calculate balances")
		return
	}

	// Calculate totals
	var totalExpenses float64
	for _, expense := range expenses {
		totalExpenses += expense.Amount
	}

	// Create CSV report
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Report header
	writer.Write([]string{"REIMBURSEMENT SUMMARY REPORT"})
	writer.Write([]string{"Team:", team.Name})
	writer.Write([]string{"Generated:", time.Now().Format("2006-01-02 15:04:05")})
	writer.Write([]string{})

	// Summary
	writer.Write([]string{"SUMMARY"})
	writer.Write([]string{"Total Expenses:", fmt.Sprintf("%.2f", totalExpenses)})
	writer.Write([]string{"Number of Expenses:", fmt.Sprintf("%d", len(expenses))})
	writer.Write([]string{})

	// Settlements needed
	writer.Write([]string{"SETTLEMENTS NEEDED"})
	writer.Write([]string{"From", "To", "Amount"})
	for _, balance := range balances.Balances {
		writer.Write([]string{
			balance.FromUser.Name + " (" + balance.FromUser.Email + ")",
			balance.ToUser.Name + " (" + balance.ToUser.Email + ")",
			fmt.Sprintf("%.2f", balance.Amount),
		})
	}
	writer.Write([]string{})

	// Member balances
	writer.Write([]string{"MEMBER BALANCES"})
	writer.Write([]string{"Name", "Email", "Owes", "Is Owed", "Net"})
	for _, member := range balances.Members {
		writer.Write([]string{
			member.User.Name,
			member.User.Email,
			fmt.Sprintf("%.2f", member.TotalOwed),
			fmt.Sprintf("%.2f", member.TotalOwing),
			fmt.Sprintf("%.2f", member.NetBalance),
		})
	}
	writer.Write([]string{})

	// Expense details
	writer.Write([]string{"EXPENSE DETAILS"})
	writer.Write([]string{"Date", "Description", "Category", "Amount", "Paid By"})
	for _, expense := range expenses {
		writer.Write([]string{
			expense.CreatedAt.Format("2006-01-02"),
			expense.Description,
			expense.Category,
			fmt.Sprintf("%.2f", expense.Amount),
			expense.PaidBy.Name,
		})
	}

	writer.Flush()

	// Set headers for file download
	filename := fmt.Sprintf("reimbursement_summary_%s_%s.csv", team.Name, time.Now().Format("20060102"))
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Write(buf.Bytes())
}
