package position

import (
	"attendance/backend/foundation/web"
	"attendance/backend/internal/auth"
	"attendance/backend/internal/entity"
	"attendance/backend/internal/pkg/repository/postgresql"
	"attendance/backend/internal/repository/postgres"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type Repository struct {
	*postgresql.Database
}

func NewRepository(database *postgresql.Database) *Repository {
	return &Repository{Database: database}
}

func (r Repository) GetById(ctx context.Context, id int) (entity.Position, error) {
	var detail entity.Position

	err := r.NewSelect().Model(&detail).Where("id = ?", id).Scan(ctx)

	return detail, err
}

func (r Repository) GetList(ctx context.Context, filter Filter) ([]GetListResponse, int, error) {
	_, err := r.CheckClaims(ctx, auth.RoleAdmin)
	if err != nil {
		return nil, 0, err
	}

	whereQuery := fmt.Sprintf(`
			WHERE 
				p.deleted_at IS NULL and d.deleted_at is null
			`)

	if filter.Search != nil {
		search := strings.Replace(*filter.Search, " ", "", -1)
		search = strings.Replace(search, "'", "''", -1)

		whereQuery += fmt.Sprintf(` AND
				(p.name ILIKE '%s')`, "%"+search+"%")
	}
	if filter.DepartmentID != nil {
		whereQuery += fmt.Sprintf(` AND p.department_id = %d`, *filter.DepartmentID)
	}

	orderQuery := "ORDER BY p.created_at desc"

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
			p.id,
			p.name,
			d.id,
			d.name
		FROM   position as p
		 JOIN department d ON p.department_id=d.id	AND d.deleted_at IS NULL
		%s %s %s %s
	`, whereQuery, orderQuery, limitQuery, offsetQuery)

	rows, err := r.QueryContext(ctx, query)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, 0, web.NewRequestError(postgres.ErrNotFound, http.StatusNotFound)
	}
	if err != nil {
		return nil, 0, web.NewRequestError(errors.Wrap(err, "selecting position"), http.StatusInternalServerError)
	}

	var list []GetListResponse

	for rows.Next() {
		var detail GetListResponse
		if err = rows.Scan(
			&detail.ID,
			&detail.Name,
			&detail.DepartmentID,
			&detail.Department); err != nil {
			return nil, 0, web.NewRequestError(errors.Wrap(err, "scanning department list"), http.StatusBadRequest)
		}

		list = append(list, detail)
	}

	countQuery := fmt.Sprintf(`
		SELECT
			count(p.id)
		FROM
		    position as p
		JOIN department d ON p.department_id=d.id	
		%s
	`, whereQuery)

	countRows, err := r.QueryContext(ctx, countQuery)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, 0, web.NewRequestError(postgres.ErrNotFound, http.StatusBadRequest)
	}
	if err != nil {
		return nil, 0, web.NewRequestError(errors.Wrap(err, "selecting positions"), http.StatusInternalServerError)
	}

	count := 0

	for countRows.Next() {
		if err = countRows.Scan(&count); err != nil {
			return nil, 0, web.NewRequestError(errors.Wrap(err, "scanning position count"), http.StatusInternalServerError)
		}
	}

	return list, count, nil
}

func (r Repository) GetDetailById(ctx context.Context, id int) (GetDetailByIdResponse, error) {
	_, err := r.CheckClaims(ctx, auth.RoleAdmin)
	if err != nil {
		return GetDetailByIdResponse{}, err
	}

	query := fmt.Sprintf(`
		SELECT
			p.id,
			p.name,
			p.department_id,
			d.name
		FROM
		    position as p
		LEFT JOIN department d ON p.department_id=d.id	
		WHERE p.deleted_at IS NULL AND p.id = %d
	`, id)

	var detail GetDetailByIdResponse

	err = r.QueryRowContext(ctx, query).Scan(
		&detail.ID,
		&detail.Name,
		&detail.DepartmentID,
		&detail.Department,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return GetDetailByIdResponse{}, web.NewRequestError(postgres.ErrNotFound, http.StatusBadRequest)
	}
	if err != nil {
		return GetDetailByIdResponse{}, web.NewRequestError(errors.Wrap(err, "selecting position  detail"), http.StatusBadRequest)
	}

	return detail, nil
}

func (r Repository) Create(ctx context.Context, request CreateRequest) (CreateResponse, error) {
	claims, err := r.CheckClaims(ctx, auth.RoleAdmin)
	if err != nil {
		return CreateResponse{}, err
	}

	if err := r.ValidateStruct(&request, "Name", "DepartmentID"); err != nil {
		return CreateResponse{}, err
	}

	var exists bool
	err = r.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM department WHERE id = ? AND deleted_at IS NULL)", request.DepartmentID).Scan(&exists)
	if err != nil {
		return CreateResponse{}, web.NewRequestError(errors.Wrap(err, "checking department existence"), http.StatusInternalServerError)
	}
	if !exists {
		return CreateResponse{}, web.NewRequestError(errors.New("無効または削除された部門ID"), http.StatusBadRequest)
	}

	var response CreateResponse

	response.Name = request.Name
	response.DepartmentID = request.DepartmentID
	response.CreatedAt = time.Now()
	response.CreatedBy = claims.UserId

	_, err = r.NewInsert().Model(&response).Returning("id").Exec(ctx, &response.ID)
	if err != nil {
		return CreateResponse{}, web.NewRequestError(errors.Wrap(err, "creating position"), http.StatusBadRequest)
	}

	return response, nil
}

func (r Repository) UpdateAll(ctx context.Context, request UpdateRequest) error {
	if err := r.ValidateStruct(&request, "ID", "Name", "DepartmentID"); err != nil {
		return err
	}

	claims, err := r.CheckClaims(ctx, auth.RoleAdmin)
	if err != nil {
		return err
	}
	var exists bool
	err = r.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM department WHERE id = ? AND deleted_at IS NULL)", request.DepartmentID).Scan(&exists)
	if err != nil {
		return web.NewRequestError(errors.Wrap(err, "checking department existence"), http.StatusInternalServerError)
	}
	if !exists {
		return web.NewRequestError(errors.New("無効または削除された部門ID"), http.StatusBadRequest)
	}

	q := r.NewUpdate().Table("position").Where("deleted_at IS NULL AND id = ?", request.ID)

	q.Set("name = ?", request.Name)
	q.Set("department_id=?", request.DepartmentID)
	q.Set("updated_at = ?", time.Now())
	q.Set("updated_by = ?", claims.UserId)

	_, err = q.Exec(ctx)
	if err != nil {
		return web.NewRequestError(errors.Wrap(err, "updating position"), http.StatusBadRequest)
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
	var exists bool
	err = r.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM department WHERE id = ? AND deleted_at IS NULL)", request.DepartmentID).Scan(&exists)
	if err != nil {
		return web.NewRequestError(errors.Wrap(err, "checking department existence"), http.StatusInternalServerError)
	}
	if !exists {
		return web.NewRequestError(errors.New("無効または削除された部門ID"), http.StatusBadRequest)
	}

	q := r.NewUpdate().Table("position").Where("deleted_at IS NULL AND id = ?", request.ID)

	if request.Name != nil {
		q.Set("name = ?", request.Name)
	}
	if request.DepartmentID != nil {
		q.Set("department_id = ?", request.DepartmentID)
	}

	q.Set("updated_at = ?", time.Now())
	q.Set("updated_by = ?", claims.UserId)

	_, err = q.Exec(ctx)
	if err != nil {
		return web.NewRequestError(errors.Wrap(err, "updating position"), http.StatusInternalServerError)
	}

	return nil
}

func (r Repository) Delete(ctx context.Context, id int) error {
	var exists bool
	err := r.DB.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT u.id 
			FROM users AS u 
			JOIN position AS p ON p.id = u.position_id 
			WHERE u.deleted_at IS NULL 
			  AND p.deleted_at IS NULL 
			  AND p.id = ?
		)
	`, id).Scan(&exists)
	if err != nil {
		return web.NewRequestError(errors.Wrap(err, "failed to check if position is in use"), http.StatusInternalServerError)
	}

	if exists {
		return web.NewRequestError(errors.New("このポジションはアクティブなユーザーに使われています。関連するユーザーを先に削除しないと、削除できません。"), http.StatusBadRequest)
	}
	return r.DeleteRow(ctx, "position", id)
}
