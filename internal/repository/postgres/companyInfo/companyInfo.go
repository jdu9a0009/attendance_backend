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

	q := r.NewUpdate().Table("company_info").Where("deleted_at IS NULL AND id = ?", request.ID)
	q.Set("company_name = ?", request.CompanyName)
	q.Set("url = ?", request.Url)
	q.Set("latitude = ?", request.Latitude)
	q.Set("longitude = ?", request.Longitude)
	q.Set("start_time = ?", request.StartTime)
	q.Set("end_time = ?", request.EndTime)
	q.Set("late_time = ?", request.LateTime)
	q.Set("over_end_time = ?", request.OverEndTime)
	q.Set("come_color=?", request.ComeColor)
	q.Set("leave_color=?", request.LeaveColor)
	q.Set("forget_time_color=?", request.ForgetTimeColor)
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
			Status: http.StatusUnauthorized,
		}
	}
	return detail, nil
}
