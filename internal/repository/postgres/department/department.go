package department

import (
	"attendance/backend/foundation/web"
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

func (r Repository) GetById(ctx context.Context, id int) (entity.Department, error) {
	var detail entity.Department

	err := r.NewSelect().Model(&detail).Where("id = ?", id).Scan(ctx)

	return detail, err
}

func (r Repository) GetList(ctx context.Context, filter Filter) ([]GetListResponse, int, int, error) {
	_, err := r.CheckClaims(ctx)
	if err != nil {
		return nil, 0, 0, err
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
	orderQuery := "ORDER BY display_number desc"
	groupQuery := "GROUP BY display_number,id"

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
			name,
			display_number,
			department_nickname
		FROM department

		%s %s %s %s %s
	`, whereQuery, groupQuery, orderQuery, limitQuery, offsetQuery)

	rows, err := r.QueryContext(ctx, query)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, 0, 0, web.NewRequestError(postgres.ErrNotFound, http.StatusBadRequest)
	}
	if err != nil {
		return nil, 0, 0, web.NewRequestError(errors.Wrap(err, "selecting department"), http.StatusBadRequest)
	}

	var list []GetListResponse

	for rows.Next() {
		var detail GetListResponse
		if err = rows.Scan(
			&detail.ID,
			&detail.Name,
			&detail.DisplayNumber,
			&detail.NickName); err != nil {
			return nil, 0, 0, web.NewRequestError(errors.Wrap(err, "scanning department list"), http.StatusBadRequest)
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
		return nil, 0, 0, web.NewRequestError(postgres.ErrNotFound, http.StatusBadRequest)
	}
	if err != nil {
		return nil, 0, 0, web.NewRequestError(errors.Wrap(err, "selecting department "), http.StatusBadRequest)
	}

	count := 0

	for countRows.Next() {
		if err = countRows.Scan(&count); err != nil {
			return nil, 0, 0, web.NewRequestError(errors.Wrap(err, "scanning department count"), http.StatusBadRequest)
		}
	}
	LastDisplayNumber := fmt.Sprintf(`SELECT COALESCE(MAX(display_number), 0) FROM department where deleted_at is null`)

	lastRows, err := r.QueryContext(ctx, LastDisplayNumber)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, 0, 0, web.NewRequestError(postgres.ErrNotFound, http.StatusBadRequest)
	}
	if err != nil {
		return nil, 0, 0, web.NewRequestError(errors.Wrap(err, "selecting last display number "), http.StatusBadRequest)
	}

	lastDisplayNumber := 0

	for lastRows.Next() {
		if err = lastRows.Scan(&lastDisplayNumber); err != nil {
			return nil, 0, 0, web.NewRequestError(errors.Wrap(err, "scanning last display number count"), http.StatusBadRequest)
		}
	}

	return list, count, lastDisplayNumber + 1, nil
}

func (r Repository) GetDetailById(ctx context.Context, id int) (GetDetailByIdResponse, error) {
	_, err := r.CheckClaims(ctx)
	if err != nil {
		return GetDetailByIdResponse{}, err
	}

	query := fmt.Sprintf(`
		SELECT
			id,
			name,
			display_number,
			department_nickname
		FROM
		    department
	
		WHERE deleted_at IS NULL AND id = %d
	`, id)

	var detail GetDetailByIdResponse

	err = r.QueryRowContext(ctx, query).Scan(
		&detail.ID,
		&detail.Name,
		&detail.DisplayNumber,
		&detail.NickName,
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

	// Validate the struct fields
	if err := r.ValidateStruct(&request, "Name", "DisplayNumber"); err != nil {
		return CreateResponse{}, err
	}

	// Trim spaces from user input fields
	*request.Name = strings.TrimSpace(*request.Name)

	// Check if any of the fields are empty
	if *request.Name == "" {
		return CreateResponse{}, web.NewRequestError(errors.New("必須項目は空欄にできません、またはスペースのみを含むことはできません。"), http.StatusBadRequest)
	}

	// Check if the department name already exists
	var exists bool
	if err := r.QueryRowContext(ctx,
		`SELECT EXISTS (SELECT 1 FROM department WHERE name = ? AND deleted_at IS NULL)`,
		*request.Name).Scan(&exists); err != nil {
		return CreateResponse{}, web.NewRequestError(errors.Wrap(err, "department name check"), http.StatusInternalServerError)
	}

	if exists {
		return CreateResponse{}, web.NewRequestError(errors.New("部門名はすでに使用されています。"), http.StatusBadRequest)
	}

	// Get the last display number from the department table
	var LastDisplayNumber int
	if err := r.QueryRowContext(ctx, `SELECT COALESCE(MAX(display_number), 0) FROM department where deleted_at is null`).Scan(&LastDisplayNumber); err != nil {
		return CreateResponse{}, web.NewRequestError(errors.Wrap(err, "fetching last display number"), http.StatusInternalServerError)
	}

	// Check if the new department's display number is valid
	if request.DisplayNumber <= 0 || request.DisplayNumber > LastDisplayNumber+1 {
		return CreateResponse{}, web.NewRequestError(
			errors.Errorf("Invalid Display Number. The maximum allowed is %d or less than this number", LastDisplayNumber+1),
			http.StatusBadRequest,
		)
	}
	if request.DisplayNumber <= LastDisplayNumber {
		_, err = r.ExecContext(ctx, `
			UPDATE department 
			SET display_number = display_number + 1 
			WHERE deleted_at IS NULL AND display_number >= ?`, request.DisplayNumber)
		if err != nil {
			return CreateResponse{}, web.NewRequestError(errors.Wrap(err, "updating display numbers"), http.StatusInternalServerError)
		}
	}

	var response CreateResponse
	response.Name = request.Name
	response.DisplayNumber = request.DisplayNumber
	response.Nickname = request.Nickname
	response.CreatedAt = time.Now()
	response.CreatedBy = claims.UserId

	_, err = r.NewInsert().Model(&response).Returning("id").Exec(ctx, &response.ID)
	if err != nil {
		return CreateResponse{}, web.NewRequestError(errors.Wrap(err, "creating department"), http.StatusBadRequest)
	}

	return response, nil
}

func (r Repository) UpdateColumns(ctx context.Context, request UpdateRequest) error {
	claims, err := r.CheckClaims(ctx)
	if err != nil {
		return err
	}
	// Validate request ID
	if err := r.ValidateStruct(&request, "ID"); err != nil {
		return err
	}

	// Trim spaces from user input fields
	*request.Name = strings.TrimSpace(*request.Name)

	// Check if any of the fields are empty
	if *request.Name == "" {
		return web.NewRequestError(errors.New("必須項目は空欄にできません、またはスペースのみを含むことはできません。"), http.StatusBadRequest)
	}

	// Check if the department name already exists

	var exists bool
	if err := r.QueryRowContext(ctx,
		`SELECT EXISTS (SELECT 1 FROM department WHERE name = ? AND id != ?  AND deleted_at IS NULL)`,
		*request.Name, request.ID).Scan(&exists); err != nil {
		return web.NewRequestError(errors.Wrap(err, "department name check"), http.StatusInternalServerError)
	}

	if exists {
		return web.NewRequestError(errors.New("部門名はすでに使用されています。"), http.StatusBadRequest)
	}

	// Get the last display number from the department table
	var LastDisplayNumber int
	if err := r.QueryRowContext(ctx, `SELECT COALESCE(MAX(display_number), 0) FROM department WHERE deleted_at IS NULL`).Scan(&LastDisplayNumber); err != nil {
		return web.NewRequestError(errors.Wrap(err, "fetching last display number"), http.StatusInternalServerError)
	}

	// Fetch the current display number for the department being updated
	var CurrentDisplayNumber int
	if err := r.QueryRowContext(ctx, `SELECT display_number FROM department WHERE id = ?`, request.ID).Scan(&CurrentDisplayNumber); err != nil {
		return web.NewRequestError(errors.Wrap(err, "fetching current display number"), http.StatusInternalServerError)
	}

	// Validate the requested display number
	if request.DisplayNumber < 1 || request.DisplayNumber > LastDisplayNumber {
		return web.NewRequestError(
			errors.Errorf("Invalid Display Number. The maximum allowed is %d or more than 0", LastDisplayNumber),
			http.StatusBadRequest,
		)
	}

	// Update the display numbers of other departments
	if request.DisplayNumber < CurrentDisplayNumber {
		// Shift departments down (increment display numbers)
		_, err = r.ExecContext(ctx, `
			UPDATE department
			SET display_number = display_number + 1
			WHERE deleted_at IS NULL 
			AND display_number >= ?
			AND display_number < ?`,
			request.DisplayNumber, CurrentDisplayNumber)
		if err != nil {
			return web.NewRequestError(errors.Wrap(err, "updating display numbers"), http.StatusInternalServerError)
		}
	} else if request.DisplayNumber > CurrentDisplayNumber {
		// Shift departments up (decrement display numbers)
		_, err = r.ExecContext(ctx, `
			UPDATE department
			SET display_number = display_number - 1
			WHERE deleted_at IS NULL 
			AND display_number <= ? 
			AND display_number > ?`,
			request.DisplayNumber, CurrentDisplayNumber)
		if err != nil {
			return web.NewRequestError(errors.Wrap(err, "updating display numbers"), http.StatusInternalServerError)
		}
	}

	// Update the current department's display number and other fields
	q := r.NewUpdate().Table("department").Where("deleted_at IS NULL AND id = ?", request.ID)
	if request.Name != nil {
		q.Set("name = ?", request.Name)
	}
	if request.DisplayNumber != 0 {
		q.Set("display_number = ?", request.DisplayNumber)
	}
	if request.Nickname != nil {
		q.Set("department_nickname = ?", request.Nickname)
	}
	fmt.Println("Request:", request.Nickname)
	q.Set("updated_at = ?", time.Now())
	q.Set("updated_by = ?", claims.UserId)

	_, err = q.Exec(ctx)
	if err != nil {
		return web.NewRequestError(errors.Wrap(err, "updating department"), http.StatusBadRequest)
	}

	return nil
}
func (r Repository) Delete(ctx context.Context, id int) error {

	var exists bool
	err := r.DB.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT u.id 
			FROM users AS u 
			JOIN department AS d ON d.id = u.department_id 
			WHERE u.deleted_at IS NULL 
			  AND d.deleted_at IS NULL 
			  AND d.id = ?
		)
	`, id).Scan(&exists)
	if err != nil {
		return web.NewRequestError(errors.Wrap(err, "failed to check if department is in use"), http.StatusInternalServerError)
	}

	if exists {
		return web.NewRequestError(errors.New("この部門はアクティブなユーザーに使われています。関連するユーザーを先に削除しないと、削除できません。"), http.StatusBadRequest)
	}
	// Fetch the current display number for the department being updated
	var CurrentDisplayNumber int
	if err := r.QueryRowContext(ctx, `SELECT display_number FROM department WHERE id = ?`, id).Scan(&CurrentDisplayNumber); err != nil {
		return web.NewRequestError(errors.Wrap(err, "fetching current display number"), http.StatusInternalServerError)
	}

	// Update the display numbers of other departments
	_, err = r.DB.ExecContext(ctx, `
		UPDATE department
		SET display_number = display_number - 1
		WHERE deleted_at IS NULL 
		AND display_number > ?
	`, CurrentDisplayNumber)
	if err != nil {
		return web.NewRequestError(errors.Wrap(err, "updating display numbers"), http.StatusInternalServerError)
	}

	return r.DeleteRow(ctx, "department", id)
}
