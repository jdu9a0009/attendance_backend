package attendance

import (
	"attendance/backend/internal/repository/postgres/attendance"
	"attendance/backend/internal/repository/postgres/companyInfo"
	"context"

	"github.com/Azure/go-autorest/autorest/date"
)

type Attendance interface {
	GetList(ctx context.Context, filter attendance.Filter) ([]attendance.GetListResponse, int, error)
	GetDetailById(ctx context.Context, id int) (attendance.GetDetailByIdResponse, error)
	GetHistoryById(ctx context.Context, employee_id string, date date.Date) ([]attendance.GetHistoryByIdResponse, int, error)
	GetOfficeLocations(ctx context.Context) ([]attendance.OfficeLocation, error)
	UpdateAll(ctx context.Context, request attendance.UpdateRequest) error
	UpdateColumns(ctx context.Context, request attendance.UpdateRequest) error
	Delete(ctx context.Context, id int) error
	GetStatistics(ctx context.Context) (attendance.GetStatisticResponse, error)
	GetPieChartStatistic(ctx context.Context) (attendance.PieChartResponse, error)
	GetBarChartStatistic(ctx context.Context) ([]attendance.BarChartResponse, error)
	GetGraphStatistic(ctx context.Context, filter attendance.GraphRequest) ([]attendance.GraphResponse, error)

	CreateByQRCode(ctx context.Context, request attendance.EnterRequest) (attendance.CreateResponse, string, error)
	CreateByPhone(ctx context.Context, request attendance.EnterRequest) (attendance.CreateResponse, error)
	ExitByPhone(ctx context.Context, request attendance.EnterRequest) (attendance.CreateResponse, error)
}
type CompanyInfo interface {
	GetAttendanceColor(ctx context.Context) (companyInfo.GetAttendanceColorResponse, error)
}
