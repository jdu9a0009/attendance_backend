package department

import (
	"attendance/backend/internal/repository/postgres/department"
	"context"
)

type Department interface {
	GetList(ctx context.Context, filter department.Filter) ([]department.GetListResponse, int, int, error)
	GetDetailById(ctx context.Context, id int) (department.GetDetailByIdResponse, error)
	Create(ctx context.Context, request department.CreateRequest) (department.CreateResponse, error)
	UpdateColumns(ctx context.Context, request department.UpdateRequest) error
	Delete(ctx context.Context, id int) error

}
