package department

import (
	"context"
	"attendance/backend/internal/repository/postgres/department"
)

type Department interface {
	GetList(ctx context.Context, filter department.Filter) ([]department.GetListResponse, int, error)
	GetDetailById(ctx context.Context, id int) (department.GetDetailByIdResponse, error)
	Create(ctx context.Context, request department.CreateRequest) (department.CreateResponse, error)
	UpdateAll(ctx context.Context, request department.UpdateRequest) error
	UpdateColumns(ctx context.Context, request department.UpdateRequest) error
	Delete(ctx context.Context, id int) error
}
