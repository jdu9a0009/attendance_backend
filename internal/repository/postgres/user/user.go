package user

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"
	"university-backend/foundation/web"
	"university-backend/internal/auth"
	"university-backend/internal/entity"
	"university-backend/internal/pkg/repository/postgresql"
	"university-backend/internal/repository/postgres"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type Repository struct {
	*postgresql.Database
}

func NewRepository(database *postgresql.Database) *Repository {
	return &Repository{Database: database}
}

func (r Repository) GetByEmployeeID(ctx context.Context, employee_id string) (entity.User, error) {
	var detail entity.User
	fmt.Println("detail;", detail)
	err := r.NewSelect().Model(&detail).Where("employee_id = ? AND deleted_at IS NULL", employee_id).Scan(ctx)
	if err != nil {
		return entity.User{}, &web.Error{
			Err:    errors.New("employee not found!"),
			Status: http.StatusUnauthorized,
		}
	}

	return detail, err
}

func (r Repository) GetList(ctx context.Context, filter Filter) ([]GetListResponse, int, error) {
	_, err := r.CheckClaims(ctx, auth.RoleEmployee)
	if err != nil {
		return nil, 0, err
	}

	whereQuery := fmt.Sprintf(`
			WHERE 
				u.deleted_at IS NULL
			`)

	if filter.Search != nil {
		search := strings.Replace(*filter.Search, " ", "", -1)
		search = strings.Replace(search, "'", "", -1)

		whereQuery += fmt.Sprintf(` AND
		u.employee_id ilike '%s' OR u.full_name ilike '%s'`, "%"+search+"%", "%"+search+"%")
	}
	orderQuery := "ORDER BY u.created_at desc"

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
			u.id,
			u.employee_id,
			u.full_name,
			u.department_id,
			d.name,
			u.position_id,
			p.name,
			u.phone,
			u.email
		FROM users u
		LEFT JOIN department d ON d.id=u.department_id
		LEFT JOIN position p ON p.id=u.position_id

		%s %s %s %s
	`, whereQuery, orderQuery, limitQuery, offsetQuery)

	rows, err := r.QueryContext(ctx, query)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, 0, web.NewRequestError(postgres.ErrNotFound, http.StatusBadRequest)
	}
	if err != nil {
		return nil, 0, web.NewRequestError(errors.Wrap(err, "selecting users"), http.StatusBadRequest)
	}

	var list []GetListResponse

	for rows.Next() {
		var detail GetListResponse
		if err = rows.Scan(
			&detail.ID,
			&detail.EmployeeID,
			&detail.FullName,
			&detail.DepartmentID,
			&detail.Department,
			&detail.PositionID,
			&detail.Position,
			&detail.Phone,
			&detail.Email); err != nil {
			return nil, 0, web.NewRequestError(errors.Wrap(err, "scanning user list"), http.StatusBadRequest)
		}

		list = append(list, detail)
	}

	countQuery := fmt.Sprintf(`
		SELECT
			count(u.id)
		FROM  users u
			%s
	`, whereQuery)

	countRows, err := r.QueryContext(ctx, countQuery)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, 0, web.NewRequestError(postgres.ErrNotFound, http.StatusBadRequest)
	}
	if err != nil {
		return nil, 0, web.NewRequestError(errors.Wrap(err, "selecting users"), http.StatusBadRequest)
	}

	count := 0

	for countRows.Next() {
		if err = countRows.Scan(&count); err != nil {
			return nil, 0, web.NewRequestError(errors.Wrap(err, "scanning user count"), http.StatusBadRequest)
		}
	}

	return list, count, nil
}

func (r Repository) GetDetailById(ctx context.Context, id int) (GetDetailByIdResponse, error) {
	_, err := r.CheckClaims(ctx, auth.RoleEmployee)
	if err != nil {
		return GetDetailByIdResponse{}, err
	}

	query := fmt.Sprintf(`
		SELECT
			u.id,
			u.employee_id,
			u.full_name,
			u.department_id,
			d.name,
			u.position_id,
			p.name,
			u.phone,
			u.email
		FROM
		    users u 
		LEFT JOIN department d ON u.department_id = d.id
		LEFT JOIN  position p ON u.position_id=p.id
		WHERE u.deleted_at IS NULL AND u.id = %d
	`, id)

	var detail GetDetailByIdResponse

	err = r.QueryRowContext(ctx, query).Scan(
		&detail.ID,
		&detail.EmployeeID,
		&detail.FullName,
		&detail.DepartmentID,
		&detail.Department,
		&detail.PositionID,
		&detail.Position,
		&detail.Phone,
		&detail.Email,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return GetDetailByIdResponse{}, web.NewRequestError(postgres.ErrNotFound, http.StatusBadRequest)
	}
	if err != nil {
		return GetDetailByIdResponse{}, web.NewRequestError(errors.Wrap(err, "selecting user detail"), http.StatusBadRequest)
	}

	return detail, nil
}

func (r Repository) Create(ctx context.Context, request CreateRequest) (CreateResponse, error) {
	claims, err := r.CheckClaims(ctx, auth.RoleEmployee)
	if err != nil {
		return CreateResponse{}, err
	}

	if err := r.ValidateStruct(&request, "EmployeeID", "Password", "FullName"); err != nil {
		return CreateResponse{}, err
	}
	rand.Seed(time.Now().UnixNano())

	userIdStatus := true
	if err := r.QueryRowContext(ctx,
		fmt.Sprintf(`SELECT 
    						CASE WHEN 
    						(SELECT id FROM users WHERE employee_id = '%s' AND deleted_at IS NULL) IS NOT NULL 
    						THEN true ELSE false END`, *request.EmployeeID)).Scan(&userIdStatus); err != nil {
		return CreateResponse{}, web.NewRequestError(errors.Wrap(err, "employee_id check"), http.StatusInternalServerError)
	}
	if userIdStatus {
		return CreateResponse{}, web.NewRequestError(errors.Wrap(errors.New(""), "employee_id is used"), http.StatusBadRequest)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(*request.Password), bcrypt.DefaultCost)
	if err != nil {
		return CreateResponse{}, web.NewRequestError(errors.Wrap(err, "hashing password"), http.StatusInternalServerError)
	}
	hashedPassword := string(hash)

	var response CreateResponse
	role := strings.ToUpper(*request.Role)
	if (role != "EMPLOYEE") && (role != "ADMIN") {
		return CreateResponse{}, web.NewRequestError(errors.New("incorrect role. role should be EMPLOYEE or ADMIN"), http.StatusBadRequest)
	}

	response.Role = &role
	response.FullName = request.FullName
	response.EmployeeID = request.EmployeeID
	response.Password = &hashedPassword
	response.DepartmentID = request.DepartmentID
	response.PositionID = request.PositionID
	response.Phone = request.Phone
	response.Email = request.Email
	response.CreatedAt = time.Now()
	response.CreatedBy = claims.UserId

	_, err = r.NewInsert().Model(&response).Returning("id").Exec(ctx, &response.ID)
	if err != nil {
		return CreateResponse{}, web.NewRequestError(errors.Wrap(err, "creating user"), http.StatusBadRequest)
	}

	response.Password = nil

	return response, nil
}

func (r Repository) UpdateAll(ctx context.Context, request UpdateRequest) error {
	claims, err := r.CheckClaims(ctx, auth.RoleEmployee)
	if err != nil {
		return err
	}

	if err := r.ValidateStruct(&request, "ID", "EmployeeID", "Role", "FullName"); err != nil {
		return err
	}
	userIdStatus := true
	if err := r.QueryRowContext(ctx, fmt.Sprintf("SELECT CASE WHEN (SELECT id FROM users WHERE employee_id = '%s' AND deleted_at IS NULL AND id != %d) IS NOT NULL THEN true ELSE false END", *request.EmployeeID, request.ID)).Scan(&userIdStatus); err != nil {
		return web.NewRequestError(errors.Wrap(err, "employee_id check"), http.StatusInternalServerError)
	}
	if userIdStatus {
		return web.NewRequestError(errors.Wrap(errors.New(""), "employee_id is used"), http.StatusInternalServerError)
	}

	q := r.NewUpdate().Table("users").Where("deleted_at IS NULL AND id = ?", request.ID)

	role := strings.ToUpper(*request.Role)
	if (role != "EMPLOYEE") && (role != "ADMIN") {
		return web.NewRequestError(errors.New("incorrect role. role should be EMPLOYEE or ADMIN"), http.StatusBadRequest)
	}

	q.Set("employee_id = ?", request.EmployeeID)
	q.Set("role = ?", role)
	q.Set("fullname = ?", request.FullName)
	q.Set("department_id=?", request.DepartmentID)
	q.Set("position_id=?", request.PositionID)
	q.Set("phone=?", request.Phone)
	q.Set("email=?", request.Email)
	q.Set("updated_at = ?", time.Now())
	q.Set("updated_by = ?", claims.UserId)

	_, err = q.Exec(ctx)
	if err != nil {
		return web.NewRequestError(errors.Wrap(err, "updating user"), http.StatusBadRequest)
	}

	return nil
}

func (r Repository) UpdateColumns(ctx context.Context, request UpdateRequest) error {
	claims, err := r.CheckClaims(ctx, auth.RoleEmployee)
	if err != nil {
		return err
	}

	if err := r.ValidateStruct(&request, "ID"); err != nil {
		return err
	}

	q := r.NewUpdate().Table("users").Where("deleted_at IS NULL AND id = ? ", request.ID)

	if request.EmployeeID != nil {
		userIdStatus := true
		if err := r.QueryRowContext(ctx, fmt.Sprintf("SELECT CASE WHEN (SELECT id FROM users WHERE employee_id = '%s' AND deleted_at IS NULL AND id != %d) IS NOT NULL THEN true ELSE false END", *request.EmployeeID, request.ID)).Scan(&userIdStatus); err != nil {
			return web.NewRequestError(errors.Wrap(err, "employee_id check"), http.StatusInternalServerError)
		}
		if userIdStatus {
			return web.NewRequestError(errors.Wrap(errors.New(""), "employee_id is used"), http.StatusInternalServerError)
		}
		q.Set("employee_id = ?", request.EmployeeID)
	}

	if request.Role != nil {
		role := strings.ToUpper(*request.Role)
		if (role != "EMPLOYEE") && (role != "ADMIN") {
			return web.NewRequestError(errors.New("incorrect role. role should be EMPLOYEE or ADMIN"), http.StatusBadRequest)
		}
		q.Set("role = ?", role)
	}

	if request.FullName != nil {
		q.Set("full_name = ?", request.FullName)
	}
	if request.DepartmentID != nil {
		q.Set("department_id = ?", request.DepartmentID)
	}
	if request.PositionID != nil {
		q.Set("position_id=?", request.PositionID)
	}
	if request.Phone != nil {
		q.Set("phone=?", request.Phone)
	}

	if request.Email != nil {
		q.Set("email=?", request.Email)
	}
	q.Set("updated_at = ?", time.Now())
	q.Set("updated_by = ?", claims.UserId)

	_, err = q.Exec(ctx)
	if err != nil {
		return web.NewRequestError(errors.Wrap(err, "updating user"), http.StatusBadRequest)
	}

	return nil
}

func (r Repository) Delete(ctx context.Context, id int) error {
	return r.DeleteRow(ctx, "users", id)
}
