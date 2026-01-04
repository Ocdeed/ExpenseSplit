package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

func New(databaseURL string) (*DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Connected to database successfully")
	return &DB{db}, nil
}

func (db *DB) RunMigrations() error {
	migrations := []string{
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`,
		
		// Users table
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			name VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		)`,

		// Teams table
		`CREATE TABLE IF NOT EXISTS teams (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			name VARCHAR(255) NOT NULL,
			created_by UUID REFERENCES users(id),
			created_at TIMESTAMP DEFAULT NOW()
		)`,

		// Team members table
		`CREATE TABLE IF NOT EXISTS team_members (
			team_id UUID REFERENCES teams(id) ON DELETE CASCADE,
			user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			role VARCHAR(50) DEFAULT 'member',
			joined_at TIMESTAMP DEFAULT NOW(),
			PRIMARY KEY (team_id, user_id)
		)`,

		// Expenses table
		`CREATE TABLE IF NOT EXISTS expenses (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			team_id UUID REFERENCES teams(id) ON DELETE CASCADE,
			paid_by UUID REFERENCES users(id),
			amount DECIMAL(10,2) NOT NULL,
			description TEXT,
			category VARCHAR(100),
			receipt_url TEXT,
			split_type VARCHAR(50) DEFAULT 'equal',
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		)`,

		// Expense splits table
		`CREATE TABLE IF NOT EXISTS expense_splits (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			expense_id UUID REFERENCES expenses(id) ON DELETE CASCADE,
			user_id UUID REFERENCES users(id),
			amount DECIMAL(10,2) NOT NULL,
			percent DECIMAL(5,2),
			is_settled BOOLEAN DEFAULT FALSE
		)`,

		// Settlements table
		`CREATE TABLE IF NOT EXISTS settlements (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			team_id UUID REFERENCES teams(id) ON DELETE CASCADE,
			from_user UUID REFERENCES users(id),
			to_user UUID REFERENCES users(id),
			amount DECIMAL(10,2) NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		)`,

		// Approvals table
		`CREATE TABLE IF NOT EXISTS approvals (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			expense_id UUID REFERENCES expenses(id) ON DELETE CASCADE,
			approved_by UUID REFERENCES users(id),
			status VARCHAR(50) DEFAULT 'pending',
			comment TEXT,
			created_at TIMESTAMP DEFAULT NOW(),
			approved_at TIMESTAMP
		)`,

		// Indexes for better query performance
		`CREATE INDEX IF NOT EXISTS idx_expenses_team_id ON expenses(team_id)`,
		`CREATE INDEX IF NOT EXISTS idx_expenses_paid_by ON expenses(paid_by)`,
		`CREATE INDEX IF NOT EXISTS idx_expense_splits_expense_id ON expense_splits(expense_id)`,
		`CREATE INDEX IF NOT EXISTS idx_expense_splits_user_id ON expense_splits(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_team_members_user_id ON team_members(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_settlements_team_id ON settlements(team_id)`,
		`CREATE INDEX IF NOT EXISTS idx_approvals_expense_id ON approvals(expense_id)`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("failed to run migration: %w\nQuery: %s", err, migration)
		}
	}

	log.Println("Database migrations completed successfully")
	return nil
}

func (db *DB) Close() error {
	return db.DB.Close()
}
