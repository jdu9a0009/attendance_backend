package auth

import (
	"context"
	"university-backend/internal/entity"
)

type User interface {
	GetByEmployeeID(ctx context.Context, login string) (entity.User, error)
}
