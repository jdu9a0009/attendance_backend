package attendance

import (
	"context"
	"university-backend/internal/repository/postgres/attendance"
)

type Attendance interface {
	GetList(ctx context.Context, filter attendance.Filter) ([]attendance.GetListResponse, int, error)
	GetDetailById(ctx context.Context, id int) (attendance.GetDetailByIdResponse, error)
	Create(ctx context.Context, request attendance.CreateRequest) (attendance.CreateResponse, error)
	UpdateAll(ctx context.Context, request attendance.UpdateRequest) error
	UpdateColumns(ctx context.Context, request attendance.UpdateRequest) error
	Delete(ctx context.Context, id int) error
	GetStatistics(ctx context.Context) (attendance.GetStatisticResponse, error)
	GetPieChartStatistic(ctx context.Context) (attendance.PieChartResponse, error)
	GetBarChartStatistic(ctx context.Context) ([]attendance.BarChartResponse, error)
	CreateByPhone(ctx context.Context, request attendance.EnterRequest) (attendance.CreateResponse, error)
	ExitByPhone(ctx context.Context, request attendance.ExitByPhoneRequest) (attendance.ExitResponse, error)
}
