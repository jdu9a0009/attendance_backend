package department

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"
	"university-backend/foundation/web"
	"university-backend/internal/entity"
	"university-backend/internal/pkg/repository/postgresql"
	"university-backend/internal/repository/postgres"

	"github.com/pkg/errors"
)

type Repository struct {
	*postgresql.Database
}

func NewRepository(database *postgresql.Database) *Repository {
	return &Repository{Database: database}
}

func (r Repository) GetById(ctx context.Context, id int) (entity.Department, error) {
	var detail entity.Department

	err := r.NewSelect().Model(&detail).Where("id = ?", id).Scan(ctx)

	return detail, err
}

func (r Repository) GetList(ctx context.Context, filter Filter) ([]GetListResponse, int, error) {
	_, err := r.CheckClaims(ctx)
	if err != nil {
		return nil, 0, err
	}

	whereQuery := fmt.Sprintf(`
			WHERE 
				deleted_at IS NULL
			`)
	if filter.Search != nil {
		search := strings.Replace(*filter.Search, " ", "", -1)
		search = strings.Replace(search, "'", "''", -1)

		whereQuery += fmt.Sprintf(` AND
				name ILIKE '%s'`, "%"+search+"%")
	}
	orderQuery := "ORDER BY created_at desc"

	var limitQuery, offsetQuery string

	if filter.Page != nil && filter.Limit != nil {
		offset := (*filter.Page - 1) * (*filter.Limit)
		filter.Offset = &offset
	}

	if filter.Limit != nil {
		limitQuery += fmt.Sprintf(" LIMIT %d", *filter.Limit)
	}

	if filter.Offset != nil {
		offsetQuery += fmt.Sprintf(" OFFSET %d", *filter.Offset)
	}

	query := fmt.Sprintf(`
		SELECT 
			id,
			name
		FROM department

		%s %s %s %s
	`, whereQuery, orderQuery, limitQuery, offsetQuery)

	rows, err := r.QueryContext(ctx, query)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, 0, web.NewRequestError(postgres.ErrNotFound, http.StatusBadRequest)
	}
	if err != nil {
		return nil, 0, web.NewRequestError(errors.Wrap(err, "selecting department"), http.StatusBadRequest)
	}

	var list []GetListResponse

	for rows.Next() {
		var detail GetListResponse
		if err = rows.Scan(
			&detail.ID,
			&detail.Name); err != nil {
			return nil, 0, web.NewRequestError(errors.Wrap(err, "scanning department list"), http.StatusBadRequest)
		}

		list = append(list, detail)
	}

	countQuery := fmt.Sprintf(`
		SELECT
			count(u.id)
		FROM  department  u
			%s
	`, whereQuery)

	countRows, err := r.QueryContext(ctx, countQuery)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, 0, web.NewRequestError(postgres.ErrNotFound, http.StatusBadRequest)
	}
	if err != nil {
		return nil, 0, web.NewRequestError(errors.Wrap(err, "selecting department "), http.StatusBadRequest)
	}

	count := 0

	for countRows.Next() {
		if err = countRows.Scan(&count); err != nil {
			return nil, 0, web.NewRequestError(errors.Wrap(err, "scanning department count"), http.StatusBadRequest)
		}
	}

	return list, count, nil
}

func (r Repository) GetDetailById(ctx context.Context, id int) (GetDetailByIdResponse, error) {
	_, err := r.CheckClaims(ctx)
	if err != nil {
		return GetDetailByIdResponse{}, err
	}

	query := fmt.Sprintf(`
		SELECT
			id,
			name
		FROM
		    department
	
		WHERE deleted_at IS NULL AND id = %d
	`, id)

	var detail GetDetailByIdResponse

	err = r.QueryRowContext(ctx, query).Scan(
		&detail.ID,
		&detail.Name,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return GetDetailByIdResponse{}, web.NewRequestError(postgres.ErrNotFound, http.StatusBadRequest)
	}
	if err != nil {
		return GetDetailByIdResponse{}, web.NewRequestError(errors.Wrap(err, "selecting department detail"), http.StatusBadRequest)
	}

	return detail, nil
}

func (r Repository) Create(ctx context.Context, request CreateRequest) (CreateResponse, error) {
	claims, err := r.CheckClaims(ctx)
	if err != nil {
		return CreateResponse{}, err
	}

	if err := r.ValidateStruct(&request, "Name"); err != nil {
		return CreateResponse{}, err
	}
	DepartmentName := true
	if err := r.QueryRowContext(ctx,
		fmt.Sprintf(`SELECT 
    						CASE WHEN 
    						(SELECT id FROM department WHERE name = '%s' AND deleted_at IS NULL) IS NOT NULL 
    						THEN true ELSE false END`, *request.Name)).Scan(&DepartmentName); err != nil {
		return CreateResponse{}, web.NewRequestError(errors.Wrap(err, "department name check"), http.StatusInternalServerError)
	}
	if DepartmentName {
		return CreateResponse{}, web.NewRequestError(errors.Wrap(errors.New(""), "Department Name is used"), http.StatusBadRequest)
	}
	var response CreateResponse

	response.Name = request.Name
	response.CreatedAt = time.Now()
	response.CreatedBy = claims.UserId

	_, err = r.NewInsert().Model(&response).Returning("id").Exec(ctx, &response.ID)
	if err != nil {
		return CreateResponse{}, web.NewRequestError(errors.Wrap(err, "creating department"), http.StatusBadRequest)
	}

	return response, nil
}

func (r Repository) UpdateAll(ctx context.Context, request UpdateRequest) error {
	if err := r.ValidateStruct(&request, "ID", "Name"); err != nil {
		return err
	}

	claims, err := r.CheckClaims(ctx)
	if err != nil {
		return err
	}
	DepartmentName := true
	if err := r.QueryRowContext(ctx,
		fmt.Sprintf(`SELECT 
    						CASE WHEN 
    						(SELECT id FROM department WHERE name = '%s' AND deleted_at IS NULL) IS NOT NULL 
    						THEN true ELSE false END`, *request.Name)).Scan(&DepartmentName); err != nil {
		return web.NewRequestError(errors.Wrap(err, "department name check"), http.StatusBadRequest)

	}
	if DepartmentName {
		return web.NewRequestError(errors.Wrap(err, "Department Name is used"), http.StatusBadRequest)

	}
	q := r.NewUpdate().Table("department").Where("deleted_at IS NULL AND id = ?", request.ID)

	q.Set("name = ?", request.Name)
	q.Set("updated_at = ?", time.Now())
	q.Set("updated_by = ?", claims.UserId)

	_, err = q.Exec(ctx)
	if err != nil {
		return web.NewRequestError(errors.Wrap(err, "updating department"), http.StatusBadRequest)
	}

	return nil
}

func (r Repository) UpdateColumns(ctx context.Context, request UpdateRequest) error {
	if err := r.ValidateStruct(&request, "ID"); err != nil {
		return err
	}

	claims, err := r.CheckClaims(ctx)
	if err != nil {
		return err
	}
	DepartmentName := true
	if err := r.QueryRowContext(ctx,
		fmt.Sprintf(`SELECT 
    						CASE WHEN 
    						(SELECT id FROM department WHERE name = '%s' AND deleted_at IS NULL) IS NOT NULL 
    						THEN true ELSE false END`, *request.Name)).Scan(&DepartmentName); err != nil {
		return web.NewRequestError(errors.Wrap(err, "department name check"), http.StatusBadRequest)

	}
	q := r.NewUpdate().Table("department").Where("deleted_at IS NULL AND id = ?", request.ID)

	if request.Name != nil {
		q.Set("name = ?", request.Name)
	}

	q.Set("updated_at = ?", time.Now())
	q.Set("updated_by = ?", claims.UserId)

	_, err = q.Exec(ctx)
	if err != nil {
		return web.NewRequestError(errors.Wrap(err, "updating department"), http.StatusBadRequest)
	}

	return nil
}

func (r Repository) Delete(ctx context.Context, id int) error {
	return r.DeleteRow(ctx, "department", id)
}
