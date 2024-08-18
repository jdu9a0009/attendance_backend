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
	_, err := r.CheckClaims(ctx, auth.RoleAdmin)
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
	_, err := r.CheckClaims(ctx, auth.RoleAdmin)
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
	claims, err := r.CheckClaims(ctx, auth.RoleAdmin)
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
	claims, err := r.CheckClaims(ctx, auth.RoleAdmin)
	if err != nil {
		return err
	}

	if err := r.ValidateStruct(&request, "ID", "EmployeeID", "Role"); err != nil {
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
	q.Set("full_name = ?", request.FullName)
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
	claims, err := r.CheckClaims(ctx, auth.RoleAdmin)
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

func (r Repository) GetMonthlyStatistics(ctx context.Context, filter StatisticRequest) (MonthlyStatisticResponse, error) {
	claims, err := r.CheckClaims(ctx)
	if err != nil {
		return MonthlyStatisticResponse{}, web.NewRequestError(errors.New("error getting user by token"), http.StatusBadRequest)
	}

	// Calculate full month dates
	monthStart := time.Date(filter.Month.Year(), filter.Month.Month(), 1, 0, 0, 0, 0, time.UTC)
	monthEnd := time.Date(filter.Month.Year(), filter.Month.Month()+1, 0, 23, 59, 59, 999999999, time.UTC)

	// Query for monthly statistics
	monthlyQuery := `
		SELECT
			SUM(CASE WHEN a.come_time <= '09:00' THEN 1 ELSE 0 END) AS early_come,
			SUM(CASE WHEN a.leave_time < '18:00' THEN 1 ELSE 0 END) AS early_leave,
			SUM(CASE WHEN u.status = 'false' THEN 1 ELSE 0 END) AS absent,
			SUM(CASE WHEN a.come_time >= '10:00' THEN 1 ELSE 0 END) AS late
		FROM attendance a
		JOIN users u ON a.employee_id = u.employee_id
		WHERE a.deleted_at IS NULL AND u.id = $1
		AND a.work_day BETWEEN $2 AND $3;
	`

	// Execute monthly query
	monthlyStmt, err := r.Prepare(monthlyQuery)
	if err != nil {
		return MonthlyStatisticResponse{}, web.NewRequestError(errors.Wrap(err, "preparing monthly query"), http.StatusInternalServerError)
	}
	defer monthlyStmt.Close()

	// Execute interval query
	Stmt, err := r.Prepare(monthlyQuery)
	if err != nil {
		return MonthlyStatisticResponse{}, web.NewRequestError(errors.Wrap(err, "preparing montly query"), http.StatusInternalServerError)
	}
	defer Stmt.Close()

	rows, err := Stmt.QueryContext(ctx, claims.UserId, monthStart, monthEnd)
	if err != nil {
		return MonthlyStatisticResponse{}, web.NewRequestError(errors.Wrap(err, "executing monthly query"), http.StatusInternalServerError)
	}
	defer rows.Close()
	var list MonthlyStatisticResponse
	for rows.Next() {
		err := rows.Scan(
			&list.EarlyCome,
			&list.EarlyLeave,
			&list.Absent,
			&list.Late,
		)
		if err != nil {
			return MonthlyStatisticResponse{}, web.NewRequestError(errors.Wrap(err, "scanning monthly statistics"), http.StatusInternalServerError)
		}

	}
	return list, nil
}

func (r Repository) GetStatistics(ctx context.Context, filter StatisticRequest) ([]StatisticResponse, error) {
	claims, err := r.CheckClaims(ctx)
	if err != nil {
		return nil, web.NewRequestError(errors.New("error getting user by token"), http.StatusBadRequest)
	}

	// Determine the start and end days based on the interval
	var startDay, endDay int
	switch filter.Interval {
	case 0:
		startDay, endDay = 1, 10
	case 1:
		startDay, endDay = 11, 20
	case 2:
		startDay, endDay = 21, 31 // Adjust for months with fewer than 31 days later
	default:
		return nil, web.NewRequestError(errors.New("invalid interval"), http.StatusBadRequest)
	}

	// Calculate start and end dates for the interval
	startDate := time.Date(filter.Month.Year(), filter.Month.Month(), startDay, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(filter.Month.Year(), filter.Month.Month(), endDay, 23, 59, 59, 999999999, time.UTC)

	// Query for interval data
	intervalQuery := `
		SELECT
			a.work_day,
COALESCE(TO_CHAR(a.come_time, 'HH24:MI'), NULL) AS come_time,
			COALESCE(TO_CHAR(a.leave_time, 'HH24:MI'), NULL) AS leave_time,
			CASE
				WHEN a.come_time IS NOT NULL AND a.leave_time IS NOT NULL THEN TO_CHAR(a.leave_time - a.come_time, 'HH24:MI')
				ELSE NULL
			END AS total_hours
		FROM attendance a
		JOIN users u ON a.employee_id = u.employee_id
		WHERE a.deleted_at IS NULL AND u.id = $1
		AND a.work_day BETWEEN $2 AND $3
		GROUP BY a.work_day, a.come_time, a.leave_time;
	`

	// Execute interval query
	intervalStmt, err := r.Prepare(intervalQuery)
	if err != nil {
		return nil, web.NewRequestError(errors.Wrap(err, "preparing interval query"), http.StatusInternalServerError)
	}
	defer intervalStmt.Close()

	rows, err := intervalStmt.QueryContext(ctx, claims.UserId, startDate, endDate)
	if err != nil {
		return nil, web.NewRequestError(errors.Wrap(err, "executing interval query"), http.StatusInternalServerError)
	}
	defer rows.Close()

	var list []StatisticResponse
	for rows.Next() {
		var detail StatisticResponse
		err := rows.Scan(
			&detail.WorkDay,
			&detail.ComeTime,
			&detail.LeaveTime,
			&detail.TotalHours,
		)
		if err != nil {
			return nil, web.NewRequestError(errors.Wrap(err, "scanning interval statistics"), http.StatusInternalServerError)
		}

		list = append(list, detail)
	}
	return list, nil
}

func (r Repository) GetEmployeeDashboard(ctx context.Context) (DashboardResponse, error) {
	claims, err := r.CheckClaims(ctx)
	if err != nil {
		return DashboardResponse{}, err
	}
	var detail DashboardResponse
	query := fmt.Sprintf(`
        SELECT
            COALESCE(TO_CHAR(a.come_time, 'HH24:MI'), NULL) AS come_time,
			COALESCE(TO_CHAR(a.leave_time, 'HH24:MI'), NULL) AS leave_time,
			CASE
				WHEN a.come_time IS NOT NULL AND a.leave_time IS NOT NULL THEN TO_CHAR(a.leave_time - a.come_time, 'HH24:MI')
				ELSE NULL
			END AS total_hours
        FROM
            attendance as a
			JOIN users u ON a.employee_id = u.employee_id

        WHERE
            a.deleted_at IS NULL
            AND u.id = %d;
	`, claims.UserId)
	err = r.QueryRowContext(ctx, query).Scan(
		&detail.ComeTime,
		&detail.LeaveTime,
		&detail.TotalHours,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return DashboardResponse{}, nil
	}
	if err != nil {
		return DashboardResponse{}, web.NewRequestError(errors.Wrap(err, "selecting user detail on dashboard"), http.StatusBadRequest)
	}
	return detail, nil
}
