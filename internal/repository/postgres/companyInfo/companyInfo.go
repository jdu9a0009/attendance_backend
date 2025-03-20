package companyInfo

import (
	"attendance/backend/foundation/web"
	"attendance/backend/internal/pkg/repository/postgresql"
	"context"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

type Repository struct {
	*postgresql.Database
}

func NewRepository(database *postgresql.Database) *Repository {
	return &Repository{Database: database}
}

func (r Repository) UpdateAll(ctx context.Context, request UpdateRequest) error {
	if err := r.ValidateStruct(&request, "company_name", "latitude", "longitude"); err != nil {
		return err
	}

	claims, err := r.CheckClaims(ctx)
	if err != nil {
		return err
	}
	radius := request.Radius
	if radius == 0 {
		radius = 3000.0
	}
	q := r.NewUpdate().Table("company_info").Where("deleted_at IS NULL AND id = ?", request.ID)
	q.Set("company_name = ?", request.CompanyName)
	q.Set("url = ?", request.Url)
	q.Set("latitude = ?", request.Latitude)
	q.Set("longitude = ?", request.Longitude)
	q.Set("radius = ?", radius)
	q.Set("start_time = ?", request.StartTime)
	q.Set("end_time = ?", request.EndTime)
	q.Set("late_time = ?", request.LateTime)
	q.Set("over_end_time = ?", request.OverEndTime)
	q.Set("come_time_color=?", request.ComeTimeColor)
	q.Set("leave_time_color=?", request.LeaveTimeColor)
	q.Set("forget_time_color=?", request.ForgetTimeColor)
	q.Set("present_color=?", request.PresentColor)
	q.Set("absent_color=?", request.AbsentColor)
	q.Set("new_present_color=?", request.NewPresentColor)
	q.Set("new_absent_color=?", request.NewAbsentColor)
	q.Set("updated_at = ?", time.Now())
	q.Set("updated_by = ?", claims.UserId)

	_, err = q.Exec(ctx)
	if err != nil {
		return web.NewRequestError(errors.Wrap(err, "updating company_info"), http.StatusBadRequest)
	}

	return nil
}

func (r Repository) GetInfo(ctx context.Context) (GetInfoResponse, error) {
	var detail GetInfoResponse
	err := r.NewSelect().
		Model(&detail).
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Limit(1).
		Scan(ctx)
	if err != nil {

		return GetInfoResponse{}, &web.Error{
			Err:    errors.New("company data not found!"),
			Status: http.StatusNotFound,
		}
	}
	return detail, nil
}

func (r Repository) GetAttendanceColor(ctx context.Context) (GetAttendanceColorResponse, error) {
	var detail GetAttendanceColorResponse
	err := r.NewSelect().
		Model(&detail).
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Limit(1).
		Scan(ctx)
	if err != nil {

		return GetAttendanceColorResponse{}, &web.Error{
			Err:    errors.New("company  attendance colors not found!"),
			Status: http.StatusUnauthorized,
		}
	}
	return detail, nil
}

func (r Repository) GetNewTableColor(ctx context.Context) (GetNewTableColorResponse, error) {
	var detail GetNewTableColorResponse
	err := r.NewSelect().
		Model(&detail).
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Limit(1).
		Scan(ctx)
	if err != nil {

		return GetNewTableColorResponse{}, &web.Error{
			Err:    errors.New("company  New Table colors not found!"),
			Status: http.StatusUnauthorized,
		}
	}
	return detail, nil
}
