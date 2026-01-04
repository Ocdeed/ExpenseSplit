package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/expensesplit/backend/internal/config"
	"github.com/expensesplit/backend/internal/database"
	"github.com/expensesplit/backend/internal/handlers"
	"github.com/expensesplit/backend/internal/middleware"
	"github.com/expensesplit/backend/internal/repository"
	"github.com/expensesplit/backend/internal/services"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database
	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Create upload directory
	if err := os.MkdirAll(cfg.UploadDir, 0755); err != nil {
		log.Fatalf("Failed to create upload directory: %v", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	teamRepo := repository.NewTeamRepository(db)
	expenseRepo := repository.NewExpenseRepository(db)
	settlementRepo := repository.NewSettlementRepository(db)
	approvalRepo := repository.NewApprovalRepository(db)

	// Initialize services
	tokenDuration, _ := time.ParseDuration(cfg.JWTExpiration)
	authService := services.NewAuthService(userRepo, cfg.JWTSecret, tokenDuration)
	teamService := services.NewTeamService(teamRepo, userRepo)
	expenseService := services.NewExpenseService(expenseRepo, teamRepo, userRepo, approvalRepo)
	balanceService := services.NewBalanceService(expenseRepo, teamRepo, userRepo, settlementRepo)
	approvalService := services.NewApprovalService(approvalRepo, expenseRepo, teamRepo)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService, userRepo)
	teamHandler := handlers.NewTeamHandler(teamService)
	expenseHandler := handlers.NewExpenseHandler(expenseService, teamService, cfg.UploadDir)
	balanceHandler := handlers.NewBalanceHandler(balanceService, teamService)
	exportHandler := handlers.NewExportHandler(expenseService, balanceService, teamService)
	approvalHandler := handlers.NewApprovalHandler(approvalService, teamService)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(authService)

	// Create router
	router := mux.NewRouter()

	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()

	// Public routes
	api.HandleFunc("/auth/register", authHandler.Register).Methods("POST")
	api.HandleFunc("/auth/login", authHandler.Login).Methods("POST")

	// Protected routes
	protected := api.PathPrefix("").Subrouter()
	protected.Use(authMiddleware.Authenticate)

	// User routes
	protected.HandleFunc("/auth/me", authHandler.GetMe).Methods("GET")

	// Team routes
	protected.HandleFunc("/teams", teamHandler.CreateTeam).Methods("POST")
	protected.HandleFunc("/teams", teamHandler.GetUserTeams).Methods("GET")
	protected.HandleFunc("/teams/{id}", teamHandler.GetTeam).Methods("GET")
	protected.HandleFunc("/teams/{id}", teamHandler.UpdateTeam).Methods("PUT")
	protected.HandleFunc("/teams/{id}", teamHandler.DeleteTeam).Methods("DELETE")
	protected.HandleFunc("/teams/{id}/members", teamHandler.GetTeamMembers).Methods("GET")
	protected.HandleFunc("/teams/{id}/members", teamHandler.AddMember).Methods("POST")
	protected.HandleFunc("/teams/{id}/members/{memberId}", teamHandler.RemoveMember).Methods("DELETE")

	// Expense routes
	protected.HandleFunc("/teams/{teamId}/expenses", expenseHandler.CreateExpense).Methods("POST")
	protected.HandleFunc("/teams/{teamId}/expenses", expenseHandler.GetTeamExpenses).Methods("GET")
	protected.HandleFunc("/teams/{teamId}/expenses/{id}", expenseHandler.GetExpense).Methods("GET")
	protected.HandleFunc("/teams/{teamId}/expenses/{id}", expenseHandler.UpdateExpense).Methods("PUT")
	protected.HandleFunc("/teams/{teamId}/expenses/{id}", expenseHandler.DeleteExpense).Methods("DELETE")
	protected.HandleFunc("/teams/{teamId}/expenses/{id}/receipt", expenseHandler.UploadReceipt).Methods("POST")

	// Approval routes
	protected.HandleFunc("/teams/{teamId}/approvals", approvalHandler.GetTeamApprovals).Methods("GET")
	protected.HandleFunc("/teams/{teamId}/approvals/{id}", approvalHandler.UpdateApprovalStatus).Methods("PUT")

	// Balance routes
	protected.HandleFunc("/teams/{teamId}/balances", balanceHandler.GetTeamBalances).Methods("GET")
	protected.HandleFunc("/teams/{teamId}/balances/me", balanceHandler.GetUserBalance).Methods("GET")
	protected.HandleFunc("/teams/{teamId}/settlements", balanceHandler.RecordSettlement).Methods("POST")

	// Export routes
	protected.HandleFunc("/teams/{teamId}/export/expenses", exportHandler.ExportExpensesCSV).Methods("GET")
	protected.HandleFunc("/teams/{teamId}/export/balances", exportHandler.ExportBalancesCSV).Methods("GET")
	protected.HandleFunc("/teams/{teamId}/export/summary", exportHandler.ExportReimbursementSummary).Methods("GET")

	// Serve uploaded files
	router.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir(cfg.UploadDir))))

	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}).Methods("GET")

	// Apply middleware
	handler := middleware.Logging(router)

	// CORS configuration
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{cfg.AllowedOrigins},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	})
	handler = c.Handler(handler)

	// Start server
	log.Printf("Server starting on port %s", cfg.ServerPort)
	log.Printf("API available at http://localhost:%s/api/v1", cfg.ServerPort)
	if err := http.ListenAndServe(":"+cfg.ServerPort, handler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
