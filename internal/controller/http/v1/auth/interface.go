package auth

import (
	"context"
	"attendance/backend/internal/entity"
)

type User interface {
	GetByEmployeeID(ctx context.Context, login string) (*entity.User, error)
	GetByEmployeeEmail(ctx context.Context, login string) (*entity.User, error)

}
