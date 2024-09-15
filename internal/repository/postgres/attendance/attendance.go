package attendance

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"
	"university-backend/foundation/web"
	"university-backend/internal/auth"
	"university-backend/internal/pkg/repository/postgresql"
	"university-backend/internal/repository/postgres"

	"github.com/Azure/go-autorest/autorest/date"
	"github.com/pkg/errors"
)

type Repository struct {
	*postgresql.Database
}

func NewRepository(database *postgresql.Database) *Repository {
	return &Repository{Database: database}
}

func (r Repository) GetList(ctx context.Context, filter Filter) ([]GetListResponse, int, error) {
	_, err := r.CheckClaims(ctx, auth.RoleAdmin, auth.RoleEmployee, auth.RoleDashboard)
	if err != nil {
		return []GetListResponse{}, 0, err
	}

	whereQuery := fmt.Sprintf(`WHERE a.deleted_at IS NULL `)

	if filter.Search != nil {
		search := strings.Replace(*filter.Search, " ", "", -1)
		search = strings.Replace(search, "'", "''", -1)
		whereQuery += fmt.Sprintf(` AND (u.employee_id ILIKE '%s' OR u.full_name ILIKE '%s')`, "%"+search+"%", "%"+search+"%")
	}

	if filter.DepartmentID != nil {
		whereQuery += fmt.Sprintf(` AND u.department_id = %d`, *filter.DepartmentID)
	}

	if filter.PositionID != nil {
		whereQuery += fmt.Sprintf(` AND u.position_id = %d`, *filter.PositionID)
	}


	if filter.Date != nil {
		date, err := time.Parse("2006-01-02", *filter.Date)
		if err != nil {
			return nil, 0, web.NewRequestError(errors.Wrap(err, "date parse"), http.StatusBadRequest)
		}
		whereQuery += fmt.Sprintf(" AND a.work_day = '%s'", date.Format("2006-01-02"))
	} else {
		today := time.Now().Format("2006-01-02")
		whereQuery += fmt.Sprintf(" AND a.work_day = '%s'", today)
	}
	groupByQuery := `GROUP BY a.id, a.employee_id, u.full_name, u.department_id, d.name, 
	u.position_id, p.name, a.work_day, a.status, a.come_time, a.leave_time`
	orderQuery := "ORDER BY a.created_at DESC"
	limitQuery, offsetQuery := "", ""

	if filter.Page != nil && filter.Limit != nil {
		offset := (*filter.Page - 1) * (*filter.Limit)
		filter.Offset = &offset
	}

	if filter.Limit != nil {
		limitQuery = fmt.Sprintf("LIMIT %d", *filter.Limit)
	}

	if filter.Offset != nil {
		offsetQuery = fmt.Sprintf("OFFSET %d", *filter.Offset)
	}

	query := fmt.Sprintf(`
		SELECT
		    a.id,
			a.employee_id,
			u.full_name,
			u.department_id,
			d.name,
			u.position_id,
			p.name,
			a.work_day,
			a.status,
			TO_CHAR(a.come_time, 'HH24:MI'),
			TO_CHAR(a.leave_time, 'HH24:MI'),
		   COALESCE(SUM(EXTRACT(EPOCH FROM (ap.leave_time - ap.come_time)) / 60)::INT, 0) AS total_minutes
		FROM attendance AS a
		LEFT JOIN users u ON a.employee_id = u.employee_id
		LEFT JOIN department d ON u.department_id = d.id
		LEFT JOIN position p ON u.position_id = p.id
		LEFT JOIN  attendance_period as ap ON ap.attendance_id=a.id
		%s %s %s %s %s
	`, whereQuery, groupByQuery, orderQuery, limitQuery, offsetQuery)

	rows, err := r.QueryContext(ctx, query)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, 0, web.NewRequestError(postgres.ErrNotFound, http.StatusNotFound)
		}
		return nil, 0, web.NewRequestError(errors.Wrap(err, "selecting attendance"), http.StatusInternalServerError)
	}
	defer rows.Close()

	var list []GetListResponse

	for rows.Next() {
		var detail GetListResponse
		var totalMinutes int

		err = rows.Scan(
			&detail.ID,
			&detail.EmployeeID,
			&detail.Fullname,
			&detail.DepartmentID,
			&detail.Department,
			&detail.PositionID,
			&detail.Position,
			&detail.WorkDay,
			&detail.Status,
			&detail.ComeTime,
			&detail.LeaveTime,
			&totalMinutes,
		)
		if err != nil {
			return nil, 0, web.NewRequestError(errors.Wrap(err, "scanning attendance list"), http.StatusBadRequest)
		}

		hours := totalMinutes / 60
		minutes := totalMinutes % 60
		totalHours := fmt.Sprintf("%02d:%02d", hours, minutes)
		detail.TotalHours = totalHours

		list = append(list, detail)
	}

	countQuery := fmt.Sprintf(`
		SELECT COUNT(a.id)
		FROM attendance AS a
		LEFT JOIN users u ON a.employee_id = u.employee_id
		LEFT JOIN department d ON u.department_id = d.id
		LEFT JOIN position p ON u.position_id = p.id
		%s
	`, whereQuery)

	countRows, err := r.QueryContext(ctx, countQuery)
	if err != nil {
		return nil, 0, web.NewRequestError(errors.Wrap(err, "counting attendance records"), http.StatusInternalServerError)
	}
	defer countRows.Close()

	var count int
	for countRows.Next() {
		if err = countRows.Scan(&count); err != nil {
			return nil, 0, web.NewRequestError(errors.Wrap(err, "scanning attendance count"), http.StatusInternalServerError)
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
			a.id,
			a.employee_id,
			u.full_name,
			u.department_id,
			d.name,
			u.position_id,
			p.name,
			a.work_day,
			a.status,
			TO_CHAR(a.come_time, 'HH24:MI'),
			TO_CHAR(a.leave_time, 'HH24:MI'),
			COALESCE(SUM(EXTRACT(EPOCH FROM (ap.leave_time - ap.come_time)) / 60)::INT, 0) AS total_minutes
		FROM attendance a
		LEFT JOIN users u ON a.employee_id = u.employee_id
		LEFT JOIN department d ON u.department_id = d.id
		LEFT JOIN position p ON u.position_id = p.id
		LEFT JOIN attendance_period  as ap ON ap.attendance_id=a.id
		WHERE a.deleted_at IS NULL AND a.id = %d 
		GROUP BY a.id, a.employee_id, u.full_name, u.department_id, d.name, 
	    u.position_id, p.name, a.work_day, a.status, a.come_time, a.leave_time
	`, id)

	var detail GetDetailByIdResponse

	var totalMinutes int
	err = r.QueryRowContext(ctx, query).Scan(
		&detail.ID,
		&detail.EmployeeID,
		&detail.Fullname,
		&detail.DepartmentID,
		&detail.Department,
		&detail.PositionID,
		&detail.Position,
		&detail.WorkDay,
		&detail.Status,
		&detail.ComeTime,
		&detail.LeaveTime,
		&totalMinutes,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return GetDetailByIdResponse{}, web.NewRequestError(postgres.ErrNotFound, http.StatusBadRequest)
	}
	if err != nil {
		return GetDetailByIdResponse{}, errors.Wrap(err, "scanning attendance details")
	}

	hours := totalMinutes / 60
	minutes := totalMinutes % 60
	totalHours := fmt.Sprintf("%02d:%02d", hours, minutes)
	detail.TotalHours = totalHours
	return detail, nil
}

func (r Repository) CreateByPhone(ctx context.Context, request EnterRequest) (CreateResponse, error) {
	claims, err := r.CheckClaims(ctx)
	if err != nil {
		return CreateResponse{}, err
	}
	if err := r.ValidateStruct(&request, "Latitude", "Longitude"); err != nil {
		return CreateResponse{}, err
	}

	existingAttendance, err := r.getExistingAttendance(ctx, request.EmployeeID)
	if err != nil {
		return CreateResponse{}, err
	}
	if existingAttendance.ComeTime != nil && existingAttendance.LeaveTime != nil {
		return r.resetLeaveTimeAndCreatePeriod(ctx, claims, existingAttendance, request.EmployeeID)
	}

	if existingAttendance.ComeTime != nil {
		return CreateResponse{}, web.NewRequestError(errors.New("まず退勤してください"), http.StatusBadRequest)
	}

	return r.createNewAttendance(ctx, claims, request)
}
func (r Repository) ExitByPhone(ctx context.Context, request EnterRequest) (CreateResponse, error) {
	claims, err := r.CheckClaims(ctx)
	if err != nil {
		return CreateResponse{}, err
	}
	if err := r.ValidateStruct(&request, "Latitude", "Longitude"); err != nil {
		return CreateResponse{}, err
	}

	existingAttendance, err := r.getExistingAttendance(ctx, request.EmployeeID)
	if err != nil {
		return CreateResponse{}, err
	}
	if existingAttendance.ComeTime != nil && existingAttendance.LeaveTime != nil {
		return CreateResponse{}, web.NewRequestError(errors.New("まず退勤してください"), http.StatusBadRequest)
	}

	if existingAttendance.ComeTime != nil {
		return r.handleExistingAttendance(ctx, claims, existingAttendance, request.EmployeeID)
	}

	return CreateResponse{}, web.NewRequestError(errors.New("まず退勤してください"), http.StatusBadRequest)
}

func (r Repository) CreateByQRCode(ctx context.Context, request EnterRequest) (CreateResponse, string, error) {
	claims, err := r.CheckClaims(ctx)
	if err != nil {
		return CreateResponse{}, "", err
	}
	if err := r.ValidateStruct(&request, "Latitude", "Longitude"); err != nil {
		return CreateResponse{}, "", err
	}

	existingAttendance, err := r.getExistingAttendance(ctx, request.EmployeeID)
	if err != nil {
		return CreateResponse{}, "", err
	}
	if existingAttendance.ComeTime != nil && existingAttendance.LeaveTime != nil {
		response, err := r.resetLeaveTimeAndCreatePeriod(ctx, claims, existingAttendance, request.EmployeeID)
		return response, "仕事へようこそ", err
	}

	if existingAttendance.ComeTime != nil {

		response, err := r.handleExistingAttendance(ctx, claims, existingAttendance, request.EmployeeID)
		return response, "無事に帰宅", err
	}

	response, err := r.createNewAttendance(ctx, claims, request)
	return response, "仕事へようこそ", err
}

func (r Repository) handleExistingAttendance(ctx context.Context, claims auth.Claims, existingAttendance CreateResponse, employeeID *string) (CreateResponse, error) {

	if existingAttendance.LeaveTime == nil {
		return r.updateLeaveTime(ctx, claims, existingAttendance, employeeID)
	}
	return CreateResponse{}, web.NewRequestError(errors.New("ERR on come_time dont hava data but leave_time has data"), http.StatusBadRequest)
}

func (r Repository) getExistingAttendance(ctx context.Context, employeeID *string) (CreateResponse, error) {
	workDay := time.Now().Format("2006-01-02")

	var existingAttendance CreateResponse
	err := r.NewSelect().
		Model(&existingAttendance).
		Where("employee_id = ? AND work_day = ?", employeeID, workDay).
		Limit(1).
		Scan(ctx)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return CreateResponse{}, web.NewRequestError(errors.Wrap(err, "checking attendance"), http.StatusBadRequest)
	}

	return existingAttendance, nil
}
func (r Repository) getExistingAttendancePeriod(ctx context.Context, attendance_id int) (AttendancePeriod, error) {
	workDay := time.Now().Format("2006-01-02")

	var existingAttendancePeriod AttendancePeriod
	err := r.NewSelect().
		Model(&existingAttendancePeriod).
		Where("attendance_id= ?  AND work_day = ?", attendance_id, workDay).
		Order("come_time DESC"). // Order by come_time in descending order
		Limit(1).
		Scan(ctx)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return AttendancePeriod{}, web.NewRequestError(errors.Wrap(err, "checking attendance_period"), http.StatusBadRequest)
	}

	return existingAttendancePeriod, nil
}
func (r Repository) updateLeaveTime(ctx context.Context, claims auth.Claims, existingAttendance CreateResponse, employeeID *string) (CreateResponse, error) {

	err := r.updateAttendanceLeaveTime(ctx, existingAttendance.ID, claims.UserId)
	if err != nil {
		return CreateResponse{}, err
	}

	err = r.updateAttendancePeriod(ctx, existingAttendance.ID)
	if err != nil {
		return CreateResponse{}, err
	}

	err = r.updateUserStatus(ctx, employeeID, false)
	if err != nil {
		return CreateResponse{}, err
	}
	currentTime := time.Now()
	workDay := currentTime.Format("2006-01-02")

	ExistingAttendance, err := r.getExistingAttendance(ctx, employeeID)
	if err != nil {
		return CreateResponse{}, err
	}
	ExistingAttendancePeriod, err := r.getExistingAttendancePeriod(ctx, existingAttendance.ID)
	if err != nil {
		return CreateResponse{}, err
	}

	return CreateResponse{
		ID:         ExistingAttendance.ID,
		EmployeeID: employeeID,
		ComeTime:   ExistingAttendancePeriod.ComeTime,
		LeaveTime:  ExistingAttendance.LeaveTime,
		WorkDay:    &workDay,
	}, nil
}

func (r Repository) resetLeaveTimeAndCreatePeriod(ctx context.Context, claims auth.Claims, existingAttendance CreateResponse, employeeID *string) (CreateResponse, error) {
	currentTime := time.Now()
	workDay := currentTime.Format("2006-01-02")
	err := r.resetAttendanceLeaveTime(ctx, existingAttendance.ID, claims.UserId)
	if err != nil {
		return CreateResponse{}, err
	}

	_, err = r.createAttendancePeriod(ctx, existingAttendance.ID)
	if err != nil {
		return CreateResponse{}, err
	}

	err = r.updateUserStatus(ctx, employeeID, true)
	if err != nil {
		return CreateResponse{}, err
	}
	ExistingAttendancePeriod, err := r.getExistingAttendancePeriod(ctx, existingAttendance.ID)
	if err != nil {
		return CreateResponse{}, err
	}
	return CreateResponse{
		ID:         existingAttendance.ID,
		EmployeeID: employeeID,
		ComeTime:   ExistingAttendancePeriod.ComeTime,
		WorkDay:    &workDay,
	}, nil
}
func (r Repository) createNewAttendance(ctx context.Context, claims auth.Claims, request EnterRequest) (CreateResponse, error) {
	currentTime := time.Now()
	workDay := currentTime.Format("2006-01-02")
	response := CreateResponse{
		EmployeeID: request.EmployeeID,
		ComeTime:   stringPointer(currentTime.Format("15:04")),
		WorkDay:    &workDay,
		CreatedAt:  currentTime,
		CreatedBy:  claims.UserId,
	}

	err := r.insertAttendance(ctx, &response)
	if err != nil {
		return CreateResponse{}, err
	}

	_, err = r.createAttendancePeriod(ctx, response.ID)
	if err != nil {
		return CreateResponse{}, err
	}

	err = r.updateUserStatus(ctx, request.EmployeeID, true)
	if err != nil {
		return CreateResponse{}, err
	}

	return response, nil
}

// Helper functions for database operations
func (r Repository) updateAttendanceLeaveTime(ctx context.Context, id int, userId int) error {
	currentTime := time.Now()
	leaveTimeStr := currentTime.Format("15:04")
	_, err := r.NewUpdate().
		Table("attendance").
		Where("deleted_at IS NULL AND id = ?", id).
		Set("leave_time = ?", leaveTimeStr).
		Set("updated_at = ?", currentTime).
		Set("updated_by = ?", userId).
		Exec(ctx)
	return err
}

func (r Repository) updateAttendancePeriod(ctx context.Context, attendanceID int) error {
	currentTime := time.Now()
	leaveTimeStr := currentTime.Format("15:04")
	_, err := r.NewUpdate().
		Table("attendance_period").
		Where(" leave_time is null and attendance_id = ? AND work_day = ?", attendanceID, currentTime.Format("2006-01-02")).
		Set("leave_time = ?", leaveTimeStr).
		Set("updated_at = ?", currentTime).
		Exec(ctx)
	return err
}

func (r Repository) resetAttendanceLeaveTime(ctx context.Context, id int, userId int) error {
	_, err := r.NewUpdate().
		Table("attendance").
		Where("id = ?", id).
		Set("leave_time = NULL").
		Set("updated_at = ?", time.Now()).
		Set("updated_by = ?", userId).
		Exec(ctx)
	return err
}

func (r Repository) createAttendancePeriod(ctx context.Context, attendanceID int) (int, error) {
	currentTime := time.Now()
	workDay := currentTime.Format("2006-01-02")
	var periods PeriodsCreate
	periods.Attendance = attendanceID
	periods.WorkDay = workDay
	periods.ComeTime = currentTime.Format("15:04")

	_, err := r.NewInsert().Model(&periods).Returning("id").Exec(ctx, &periods.ID)
	return periods.ID, err
}

func (r Repository) insertAttendance(ctx context.Context, response *CreateResponse) error {
	_, err := r.NewInsert().
		Model(response).
		Column("employee_id", "work_day", "come_time", "leave_time", "created_at", "created_by").
		Returning("id").
		Exec(ctx, &response.ID)
	return err
}

func (r Repository) updateUserStatus(ctx context.Context, employeeID *string, status bool) error {
	_, err := r.NewUpdate().
		Table("users").
		Where("deleted_at IS NULL AND employee_id = ?", employeeID).
		Set("status = ?", status).
		Exec(ctx)
	return err
}

func stringPointer(s string) *string {
	return &s
}

func (r Repository) UpdateAll(ctx context.Context, request UpdateRequest) error {
	if err := r.ValidateStruct(&request, "ID"); err != nil {
		return err
	}

	claims, err := r.CheckClaims(ctx)
	if err != nil {
		return err
	}

	// Use a valid date for time parsing
	comeTime, err := time.Parse("2006-01-02 15:04", "1970-01-01 "+request.ComeTime)
	if err != nil {
		return web.NewRequestError(errors.Wrap(err, "parsing come time"), http.StatusBadRequest)
	}

	var leaveTime *time.Time
	if request.LeaveTime != "" {
		t, err := time.Parse("2006-01-02 15:04", "1970-01-01 "+request.LeaveTime)
		if err != nil {
			return web.NewRequestError(errors.Wrap(err, "parsing leave time"), http.StatusBadRequest)
		}
		leaveTime = &t
	}

	q := r.NewUpdate().Table("attendance").Where("deleted_at IS NULL AND id = ?", request.ID)
	q.Set("come_time=?", comeTime.Format("15:04"))
	if leaveTime != nil {
		q.Set("leave_time=?", leaveTime.Format("15:04"))
	}
	q.Set("work_day=?", request.WorkDay)
	q.Set("updated_at = ?", time.Now())
	q.Set("updated_by = ?", claims.UserId)

	_, err = q.Exec(ctx)
	if err != nil {
		return web.NewRequestError(errors.Wrap(err, "updating attendance"), http.StatusBadRequest)
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

	q := r.NewUpdate().Table("attendance").Where("deleted_at IS NULL AND id = ?", request.ID)

	if request.ComeTime != "" {
		q.Set("come_time = ?", request.ComeTime)
	}
	if request.LeaveTime != "" {
		q.Set("leave_time = ?", request.LeaveTime)
	}
	if request.WorkDay != "" {
		q.Set("work_day = ?", request.WorkDay)
	}

	q.Set("updated_at = ?", time.Now())
	q.Set("updated_by = ?", claims.UserId)

	_, err = q.Exec(ctx)
	if err != nil {
		return web.NewRequestError(errors.Wrap(err, "updating attendance"), http.StatusBadRequest)
	}

	return nil
}

func (r Repository) Delete(ctx context.Context, id int) error {
	return r.DeleteRow(ctx, "attendance", id)
}

func (r Repository) GetStatistics(ctx context.Context) (GetStatisticResponse, error) {
	var response GetStatisticResponse
	workDay := time.Now().Format("2006-01-02")

	query := `
SELECT
    (SELECT COUNT(DISTINCT employee_id) FROM users WHERE role='EMPLOYEE' AND deleted_at IS NULL) AS total_employee,
    (SELECT COUNT(employee_id) FROM attendance WHERE come_time >= '09:00' AND come_time < '10:00' AND deleted_at IS NULL AND work_day = ?) AS on_time,
    (SELECT COUNT(DISTINCT u.employee_id) FROM users u LEFT JOIN attendance a ON u.employee_id = a.employee_id
     AND a.work_day = ? WHERE role='EMPLOYEE' AND u.deleted_at IS NULL AND a.employee_id IS NULL) AS absent,
    (SELECT COUNT(employee_id) FROM attendance WHERE come_time >= '10:00' AND deleted_at IS NULL AND work_day =?) AS late_arrival,
    (SELECT COUNT(employee_id) FROM attendance WHERE leave_time < '18:00' AND deleted_at IS NULL AND work_day = ?) AS early_departures,
    (SELECT COUNT(employee_id) FROM attendance WHERE come_time < '09:00' AND deleted_at IS NULL AND work_day = ?) AS early_come;
	`

	err := r.DB.QueryRowContext(ctx, query, workDay, workDay, workDay, workDay, workDay).Scan(
		&response.TotalEmployee,
		&response.OnTime,
		&response.Absent,
		&response.LateArrival,
		&response.EarlyDepartures,
		&response.EarlyCome,
	)
	if err != nil {
		return GetStatisticResponse{}, web.NewRequestError(errors.Wrap(err, "fetching statistics"), http.StatusBadRequest)
	}

	return response, nil
}
func (r Repository) GetPieChartStatistic(ctx context.Context) (PieChartResponse, error) {
	workDay := time.Now().Format("2006-01-02")

	query := `
WITH today_attendance AS (
    SELECT
        COUNT(DISTINCT a.employee_id) FILTER (WHERE a.work_day = ?) AS come_count,
        COUNT(DISTINCT u.employee_id) AS total_count,
        COUNT(u.employee_id) FILTER (WHERE a.employee_id IS NULL) AS absent_count
    FROM users u
    LEFT JOIN attendance a ON a.employee_id = u.employee_id AND a.work_day = ?
    WHERE u.status = 'false' AND u.deleted_at IS NULL
)
SELECT
    COALESCE(ROUND(100.0 * come_count / GREATEST(1, total_count), 2), 0) AS come_percentage,
    COALESCE(ROUND(100.0 * absent_count / GREATEST(1, total_count), 2), 0) AS absent_percentage
FROM today_attendance;
`

	var detail PieChartResponse
	var comePercentage, absentPercentage float64

	row := r.QueryRowContext(ctx, query, workDay, workDay)
	err := row.Scan(&comePercentage, &absentPercentage)
	if err != nil {
		return PieChartResponse{}, web.NewRequestError(errors.Wrap(err, "response pie chart data not found"), http.StatusBadRequest)
	}

	detail.Come = Int(int(comePercentage))
	detail.Absent = Int(int(absentPercentage))

	return detail, err
}

func Int(i int) *int {
	return &i
}

func (r Repository) GetBarChartStatistic(ctx context.Context) ([]BarChartResponse, error) {
	workDay := time.Now().Format("2006-01-02")
	fmt.Println("Workday:", workDay)
	query := `
    WITH today_attendance AS (
        SELECT
            COUNT(DISTINCT a.employee_id) FILTER (WHERE a.work_day = ?) AS come_count,
            COUNT(DISTINCT u.employee_id) AS total_count,
            u.department_id
        FROM department d
        LEFT JOIN users u ON d.id = u.department_id AND u.deleted_at IS NULL
        LEFT JOIN attendance a ON a.employee_id = u.employee_id AND a.deleted_at IS NULL
        WHERE d.deleted_at IS NULL
        GROUP BY u.department_id
    )
    SELECT
        d.name AS department,
        COALESCE(ROUND(100.0 * come_count / GREATEST(1, total_count), 2), 0) AS percentage
    FROM department d
    LEFT JOIN today_attendance ON d.id = today_attendance.department_id
    WHERE d.deleted_at IS NULL;
    `

	rows, err := r.DB.QueryContext(ctx, query, workDay)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []BarChartResponse

	for rows.Next() {
		var result BarChartResponse
		if err := rows.Scan(&result.Department, &result.Percentage); err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}
func (r Repository) GetGraphStatistic(ctx context.Context, filter GraphRequest) ([]GraphResponse, error) {
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

	startDate := time.Date(filter.Month.Year(), filter.Month.Month(), startDay, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(filter.Month.Year(), filter.Month.Month(), endDay, 23, 59, 59, 999999999, time.UTC)

	query := `
WITH today_attendance AS (
    SELECT
        a.work_day,  -- work_day is of type DATE in the database
        COUNT(DISTINCT a.employee_id) AS come_count,
        (SELECT COUNT(DISTINCT employee_id) FROM users WHERE deleted_at IS NULL) AS total_count
    FROM attendance a
    LEFT JOIN users u ON a.employee_id = u.employee_id
    WHERE a.deleted_at IS NULL
        AND a.work_day BETWEEN $1 AND $2
    GROUP BY a.work_day
)
SELECT
    work_day,
    COALESCE(ROUND(100.0 * come_count / GREATEST(1, total_count), 2), 0) AS percentage
FROM today_attendance;
    `

	stmt, err := r.Prepare(query)
	if err != nil {
		return nil, web.NewRequestError(errors.Wrap(err, "selecting attendance query"), http.StatusInternalServerError)
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, startDate, endDate)
	if err != nil {
		return nil, web.NewRequestError(errors.Wrap(err, "selecting attendance filter"), http.StatusInternalServerError)
	}
	defer rows.Close()

	var list []GraphResponse
	attendanceMap := make(map[string]float64)

	for rows.Next() {
		var workDayString string
		var percentage float64
		if err = rows.Scan(&workDayString, &percentage); err != nil {
			return nil, web.NewRequestError(errors.Wrap(err, "scanning Graph response"), http.StatusBadRequest)
		}

		attendanceMap[workDayString] = percentage
	}

	for day := startDay; day <= endDay; day++ {
		workDay := time.Date(filter.Month.Year(), filter.Month.Month(), day, 0, 0, 0, 0, time.UTC)
		workDayString := workDay.Format("2006-01-02")
		percentage, exists := attendanceMap[workDayString]

		if !exists {
			percentage = 0
		}

		parsedWorkDay, _ := date.ParseDate(workDayString) // Convert to *date.Date

		list = append(list, GraphResponse{
			WorkDay:    &parsedWorkDay,
			Percentage: percentage,
		})
	}

	return list, nil
}
