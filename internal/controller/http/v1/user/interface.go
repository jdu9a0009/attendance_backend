package user

import (
	"context"
	"university-backend/internal/repository/postgres/user"
)

type User interface {
	GetList(ctx context.Context, filter user.Filter) ([]user.GetListResponse, int, error)
	GetStatistics(ctx context.Context, filter user.StatisticRequest) ([]user.StatisticResponse, error)
	GetMonthlyStatistics(ctx context.Context, filter user.MonthlyStatisticRequest) (user.MonthlyStatisticResponse, error)
	GetEmployeeDashboard(ctx context.Context) (user.DashboardResponse, error)
	GetDetailById(ctx context.Context, id int) (user.GetDetailByIdResponse, error)
	Create(ctx context.Context, request user.CreateRequest) (user.CreateResponse, error)
	UpdateAll(ctx context.Context, request user.UpdateRequest) error
	UpdateColumns(ctx context.Context, request user.UpdateRequest) error
	Delete(ctx context.Context, id int) error
}
