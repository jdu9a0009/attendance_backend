package companyInfo

import (
	"context"
	"fmt"
	"net/http"
	"time"
	"university-backend/foundation/web"
	"university-backend/internal/pkg/repository/postgresql"

	"github.com/pkg/errors"
)

type Repository struct {
	*postgresql.Database
}

func NewRepository(database *postgresql.Database) *Repository {
	return &Repository{Database: database}
}

func (r Repository) Create(ctx context.Context, request CreateRequest) (CreateResponse, error) {
	claims, err := r.CheckClaims(ctx)
	if err != nil {
		return CreateResponse{}, err
	}
	var response CreateResponse
	fmt.Println("Request", request)
	response.CompanyName = request.CompanyName
	response.Url = request.Url
	response.Latitude = request.Latitude
	response.Longitude = request.Longitude
	response.CreatedAt = time.Now()
	response.CreatedBy = claims.UserId

	_, err = r.NewInsert().Model(&response).Returning("id").Exec(ctx, &response.ID)
	if err != nil {
		return CreateResponse{}, web.NewRequestError(errors.Wrap(err, "creating company info"), http.StatusBadRequest)
	}

	return response, nil
}

func (r Repository) UpdateColumns(ctx context.Context, request UpdateRequest) error {
	if err := r.ValidateStruct(&request, "ID"); err != nil {
		return err
	}

	claims, err := r.CheckClaims(ctx)
	if err != nil {
		return err
	}

	q := r.NewUpdate().Table("company_info").Where("deleted_at IS NULL AND id = ?", request.ID)

	if request.StartTime != "" {
		q.Set("start_time = ?", request.StartTime)
	}
	if request.EndTime != "" {
		q.Set("end_time = ?", request.EndTime)
	}
	if request.LateTime != "" {
		q.Set("late_time = ?", request.LateTime)
	}
	if request.OverEndTime != "" {
		q.Set("over_end_time = ?", request.OverEndTime)
	}

	q.Set("updated_at = ?", time.Now())
	q.Set("updated_by = ?", claims.UserId)

	_, err = q.Exec(ctx)
	if err != nil {
		return web.NewRequestError(errors.Wrap(err, "updating company info times "), http.StatusInternalServerError)
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
