package position

import (
	"context"
	"attendance/backend/internal/repository/postgres/position"
)

type Position interface {
	GetList(ctx context.Context, filter position.Filter) ([]position.GetListResponse, int, error)
	GetDetailById(ctx context.Context, id int) (position.GetDetailByIdResponse, error)
	Create(ctx context.Context, request position.CreateRequest) (position.CreateResponse, error)
	UpdateAll(ctx context.Context, request position.UpdateRequest) error
	UpdateColumns(ctx context.Context, request position.UpdateRequest) error
	Delete(ctx context.Context, id int) error
}
