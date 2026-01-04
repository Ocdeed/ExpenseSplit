package handlers

import (
	"context"

	"github.com/expensesplit/backend/internal/appcontext"
	"github.com/google/uuid"
)

func GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	return appcontext.GetUserID(ctx)
}

func GetUserEmailFromContext(ctx context.Context) (string, bool) {
	return appcontext.GetUserEmail(ctx)
}
