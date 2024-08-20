package attendance

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

	"github.com/Azure/go-autorest/autorest/date"
	"github.com/pkg/errors"
)

type Repository struct {
	*postgresql.Database
}

func NewRepository(database *postgresql.Database) *Repository {
	return &Repository{Database: database}
}

func (r Repository) GetById(ctx context.Context, id int) (entity.Attendance, error) {
	var detail entity.Attendance

	err := r.NewSelect().Model(&detail).Where("id = ?", id).Scan(ctx)

	return detail, err
}

func (r Repository) GetList(ctx context.Context, filter Filter) ([]GetListResponse, int, error) {
	_, err := r.CheckClaims(ctx)
	if err != nil {
		return nil, 0, err
	}

	whereQuery := fmt.Sprintf(`WHERE a.deleted_at IS NULL`)

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

	if filter.Status != nil {
		statusValue := "false"
		if *filter.Status {
			statusValue = "true"
		}
		whereQuery += fmt.Sprintf(" AND a.status = %s", statusValue)
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
			a.come_time,
			a.leave_time
		FROM attendance AS a
		LEFT JOIN users u ON a.employee_id = u.employee_id
		LEFT JOIN department d ON u.department_id = d.id
		LEFT JOIN position p ON u.position_id = p.id
		%s %s %s %s
	`, whereQuery, orderQuery, limitQuery, offsetQuery)

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
		var workDayString string
		var comeTimeBytes, leaveTimeBytes []byte

		err = rows.Scan(
			&detail.ID,
			&detail.EmployeeID,
			&detail.Fullname,
			&detail.DepartmentID,
			&detail.Department,
			&detail.PositionID,
			&detail.Position,
			&workDayString,
			&detail.Status,
			&comeTimeBytes,
			&leaveTimeBytes,
		)
		if err != nil {
			return nil, 0, web.NewRequestError(errors.Wrap(err, "scanning attendance list"), http.StatusBadRequest)
		}

		workDay, err := date.ParseDate(workDayString)
		if err != nil {
			return nil, 0, web.NewRequestError(errors.Wrap(err, "converting work_day to date.Date"), http.StatusBadRequest)
		}
		detail.WorkDay = &workDay

		// Convert byte array to time.Time
		comeTimeStr := string(comeTimeBytes)
		comeTime, err := time.Parse("15:04:05", comeTimeStr)
		if err != nil {
			return nil, 0, web.NewRequestError(errors.Wrap(err, "converting come_time to time.Time"), http.StatusBadRequest)
		}
		detail.ComeTime = &comeTime

		if leaveTimeBytes != nil {
			leaveTimeStr := string(leaveTimeBytes)
			leaveTime, err := time.Parse("15:04:05", leaveTimeStr)
			if err != nil {
				return nil, 0, web.NewRequestError(errors.Wrap(err, "converting leave_time to time.Time"), http.StatusBadRequest)
			}
			detail.LeaveTime = &leaveTime

			// Calculate total working hours
			totalHours, err := r.calculateTotalHours(ctx, detail.EmployeeID, *detail.WorkDay)
			if err != nil {
				return nil, 0, web.NewRequestError(errors.Wrap(err, "calculating total hours"), http.StatusBadRequest)
			}
			detail.TotalHours = totalHours
		} else {
			detail.TotalHours = ""
		}

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

// calculateTotalHours calculates the total hours worked by the employee based on attendance_period records.
func (r Repository) calculateTotalHours(ctx context.Context, employeeID *string, workDay date.Date) (string, error) {
	var totalMinutes int
	query := `
		SELECT COALESCE(SUM(EXTRACT(EPOCH FROM (leave_time - come_time)) / 60)::INT, 0) AS total_minutes
		FROM attendance_period
		WHERE attendance_id = (
			SELECT id FROM attendance WHERE employee_id = ? AND work_day = ?
		)`

	err := r.QueryRowContext(ctx, query, employeeID, workDay).Scan(&totalMinutes)
	if err != nil {
		return "", errors.Wrap(err, "scanning total minutes")
	}

	hours := totalMinutes / 60
	minutes := totalMinutes % 60
	totalHours := fmt.Sprintf("%02d:%02d", hours, minutes)

	return totalHours, nil
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
			a.come_time,
			a.leave_time
		FROM attendance a
		LEFT JOIN users u ON a.employee_id = u.employee_id
		LEFT JOIN department d ON u.department_id = d.id
		LEFT JOIN position p ON u.position_id = p.id
		WHERE a.deleted_at IS NULL AND a.id = %d
	`, id)

	var detail GetDetailByIdResponse

	var workDayString string
	var leaveTimeBytes []byte
	var comeTimeBytes []byte
	err = r.QueryRowContext(ctx, query).Scan(
		&detail.ID,
		&detail.EmployeeID,
		&detail.Fullname,
		&detail.DepartmentID,
		&detail.Department,
		&detail.PositionID,
		&detail.Position,
		&workDayString,
		&detail.Status,
		&comeTimeBytes,
		&leaveTimeBytes,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return GetDetailByIdResponse{}, web.NewRequestError(postgres.ErrNotFound, http.StatusBadRequest)
	}
	if err != nil {
		return GetDetailByIdResponse{}, errors.Wrap(err, "scanning attendance details")
	}

	// Parse work_day
	workDay, err := date.ParseDate(workDayString)
	if err != nil {
		return GetDetailByIdResponse{}, web.NewRequestError(errors.Wrap(err, "parsing work_day"), http.StatusBadRequest)
	}
	detail.WorkDay = &workDay

	// Convert come_time
	comeTimeStr := string(comeTimeBytes)
	comeTime, err := time.Parse("15:04:05", comeTimeStr)
	if err != nil {
		return GetDetailByIdResponse{}, web.NewRequestError(errors.Wrap(err, "parsing come_time"), http.StatusBadRequest)
	}
	comeTimeFormatted := comeTime.Format("15:04")
	detail.ComeTime = &comeTimeFormatted

	// Convert leave_time
	if leaveTimeBytes != nil {
		leaveTimeStr := string(leaveTimeBytes)
		leaveTime, err := time.Parse("15:04:05", leaveTimeStr)
		if err != nil {
			return GetDetailByIdResponse{}, web.NewRequestError(errors.Wrap(err, "parsing leave_time"), http.StatusBadRequest)
		}
		leaveTimeFormatted := leaveTime.Format("15:04")
		detail.LeaveTime = &leaveTimeFormatted

		// Calculate total working hours
		totalHours, err := r.calculateTotalHours(ctx, detail.EmployeeID, *detail.WorkDay)
		if err != nil {
			return GetDetailByIdResponse{}, web.NewRequestError(errors.Wrap(err, "calculating total_hours"), http.StatusBadRequest)
		}
		detail.TotalHours = totalHours
	} else {
		detail.TotalHours = ""
	}

	return detail, nil
}

func (r Repository) Create(ctx context.Context, request CreateRequest) (CreateResponse, error) {
	claims, err := r.CheckClaims(ctx)
	if err != nil {
		return CreateResponse{}, err
	}

	if err := r.ValidateStruct(&request, "EmployeeID"); err != nil {
		return CreateResponse{}, err
	}

	var response CreateResponse

	response.ComeTime = time.Now().Format("15:04")
	response.WorkDay = time.Now().Format("2006-01-02")
	response.CreatedAt = time.Now()
	response.CreatedBy = claims.UserId

	_, err = r.NewInsert().Model(&response).Returning("id").Exec(ctx, &response.ID)
	if err != nil {
		return CreateResponse{}, web.NewRequestError(errors.Wrap(err, "creating attendance by qr code"), http.StatusBadRequest)
	}

	return response, nil
}

func (r Repository) CreateByPhone(ctx context.Context, request EnterRequest) (CreateResponse, error) {
	claims, err := r.CheckClaims(ctx)
	if err != nil {
		return CreateResponse{}, err
	}

	if err := r.ValidateStruct(&request, "Latitude", "Longitude"); err != nil {
		return CreateResponse{}, err
	}

	var response CreateResponse
	currentTime := time.Now()
	workDay := currentTime.Format("2006-01-02")

	// Check if there is existing attendance data for the day
	var existingAttendance ExitResponse
	err = r.NewSelect().
		Model(&existingAttendance).
		Where("employee_id = ? AND work_day = ?", request.EmployeeID, workDay).
		Limit(1).
		Scan(ctx)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return CreateResponse{}, web.NewRequestError(errors.Wrap(err, "checking attendance"), http.StatusBadRequest)
	}

	if existingAttendance.ComeTime != "" {
		if existingAttendance.LeaveTime == nil {
			// If come_time exists and leave_time is nil, return an error message
			return CreateResponse{}, web.NewRequestError(errors.New("you have already clicked the button, please click the leave button"), http.StatusBadRequest)
		}

		// Update existing record if come_time exists and leave_time is not nil
		existingAttendance.LeaveTime = nil
		existingAttendance.UpdatedAt = currentTime
		existingAttendance.UpdatedBy = claims.UserId

		_, err = r.NewUpdate().
			Model(&existingAttendance).
			Where("id = ?", existingAttendance.ID).
			Exec(ctx)

		if err != nil {
			return CreateResponse{}, web.NewRequestError(errors.Wrap(err, "updating attendance by phone"), http.StatusBadRequest)
		}

		// Update or create the attendance_period
		var periods PeriodsCreate
		periods.Attendance = existingAttendance.ID
		periods.WorkDay = currentTime.Format("2006-01-02")
		periods.ComeTime = currentTime.Format("15:04")

		_, err = r.NewInsert().Model(&periods).Returning("id").Exec(ctx, &periods.ID)
		if err != nil {
			return CreateResponse{}, web.NewRequestError(errors.Wrap(err, "creating attendance_period by phone"), http.StatusBadRequest)
		}

		// Populate the response with updated data
		response.ID = existingAttendance.ID
		response.EmployeeID = request.EmployeeID
		response.ComeTime = existingAttendance.ComeTime
		response.WorkDay = workDay
	} else {
		// Prepare response data for a new entry
		response.EmployeeID = request.EmployeeID
		response.ComeTime = currentTime.Format("15:04")
		response.WorkDay = currentTime.Format("2006-01-02")
		response.CreatedAt = currentTime
		response.CreatedBy = claims.UserId

		_, err = r.NewInsert().Model(&response).Returning("id").Exec(ctx, &response.ID)
		if err != nil {
			return CreateResponse{}, web.NewRequestError(errors.Wrap(err, "creating attendance by phone"), http.StatusBadRequest)
		}

		// Create the attendance_period
		var periods PeriodsCreate
		periods.Attendance = response.ID
		periods.WorkDay = currentTime.Format("2006-01-02")
		periods.ComeTime = currentTime.Format("15:04")

		_, err = r.NewInsert().Model(&periods).Returning("id").Exec(ctx, &periods.ID)
		if err != nil {
			return CreateResponse{}, web.NewRequestError(errors.Wrap(err, "creating attendance_period by phone"), http.StatusBadRequest)
		}
	}

	// Update user's status to true
	_, err = r.NewUpdate().
		Table("users").
		Where("deleted_at IS NULL AND employee_id = ?", request.EmployeeID).
		Set("status = true").
		Exec(ctx)
	if err != nil {
		return CreateResponse{}, web.NewRequestError(errors.Wrap(err, "updating user's status by phone when entering"), http.StatusBadRequest)
	}

	return response, nil
}

func (r Repository) ExitByPhone(ctx context.Context, request EnterRequest) (ExitResponse, error) {
	claims, err := r.CheckClaims(ctx)
	if err != nil {
		return ExitResponse{}, err
	}

	if err := r.ValidateStruct(&request, "Latitude", "Longitude"); err != nil {
		return ExitResponse{}, err
	}

	var response ExitResponse
	currentTime := time.Now()
	workDay := currentTime.Format("2006-01-02")
	leaveTimeStr := currentTime.Format("15:04")

	// Check if there is existing attendance data for the day
	var existingAttendance ExitResponse
	err = r.NewSelect().
		Model(&existingAttendance).
		Where("employee_id = ? AND work_day = ?", request.EmployeeID, workDay).
		Limit(1).
		Scan(ctx)
	if err != nil {
		return ExitResponse{}, web.NewRequestError(errors.New("please click the come button first"), http.StatusBadRequest)
	}

	// Case 1: If no come_time exists, prompt to click come button first
	if existingAttendance.ComeTime == "" {
		return ExitResponse{}, web.NewRequestError(errors.New("please click the come button first"), http.StatusBadRequest)
	}

	// Case 2: If come_time exists but leave_time is nil, update the tables
	if existingAttendance.ComeTime != "" && existingAttendance.LeaveTime == nil {
		// Update attendance table
		_, err = r.NewUpdate().
			Table("attendance").
			Where("deleted_at IS NULL AND id = ?", existingAttendance.ID).
			Set("leave_time = ?", leaveTimeStr).
			Set("updated_at = ?", currentTime).
			Set("updated_by = ?", claims.UserId).
			Set("status = ?", false).
			Exec(ctx)
		if err != nil {
			return ExitResponse{}, web.NewRequestError(errors.Wrap(err, "updating attendance"), http.StatusBadRequest)
		}

		// Update attendance_period table
		_, err = r.NewUpdate().
			Table("attendance_period").
			Where("updated_at is null and attendance_id = ? AND work_day = ?", existingAttendance.ID, workDay).
			Set("leave_time = ?", leaveTimeStr).
			Set("updated_at = ?", currentTime).
			Exec(ctx)
		if err != nil {
			return ExitResponse{}, web.NewRequestError(errors.Wrap(err, "updating attendance_period"), http.StatusBadRequest)
		}

		// Update user's status to false
		_, err = r.NewUpdate().
			Table("users").
			Where("deleted_at IS NULL AND employee_id = ?", request.EmployeeID).
			Set("status = false").
			Exec(ctx)
		if err != nil {
			return ExitResponse{}, web.NewRequestError(errors.Wrap(err, "updating user's status by phone when exiting"), http.StatusBadRequest)
		}

		// Populate the response with updated data
		response.ID = existingAttendance.ID
		response.EmployeeID = existingAttendance.EmployeeID
		response.WorkDay = workDay
		response.ComeTime = existingAttendance.ComeTime
		response.LeaveTime = &leaveTimeStr
	}

	// Case 3: If both come_time and leave_time exist, prompt to click the come button first
	if existingAttendance.ComeTime != "" && existingAttendance.LeaveTime != nil {
		return ExitResponse{}, web.NewRequestError(errors.New("you have already recorded an exit for today. Please click the come button first"), http.StatusBadRequest)
	}

	return response, nil
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
	q.Set("come_time=?", comeTime.Format("15:04:00"))
	if leaveTime != nil {
		q.Set("leave_time=?", leaveTime.Format("15:04:00"))
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

	query := `
SELECT
    (SELECT COUNT(DISTINCT employee_id) FROM users WHERE deleted_at IS NULL) AS total_employee,
    (SELECT COUNT(employee_id) FROM attendance WHERE come_time >= '09:00' AND come_time < '10:00' AND deleted_at IS NULL AND work_day = CURRENT_DATE) AS on_time,
    (SELECT COUNT(u.employee_id) FROM users u LEFT JOIN attendance a ON a.employee_id = u.employee_id AND a.work_day = CURRENT_DATE WHERE u.status = 'false' AND u.deleted_at IS NULL AND a.employee_id IS NULL) AS absent,
    (SELECT COUNT(employee_id) FROM attendance WHERE come_time >= '10:00' AND deleted_at IS NULL AND work_day = CURRENT_DATE) AS late_arrival,
    (SELECT COUNT(employee_id) FROM attendance WHERE leave_time < '18:00' AND deleted_at IS NULL AND work_day = CURRENT_DATE) AS early_departures,
    (SELECT COUNT(employee_id) FROM attendance WHERE come_time < '09:00' AND deleted_at IS NULL AND work_day = CURRENT_DATE) AS early_come;
	`

	err := r.DB.QueryRowContext(ctx, query).Scan(
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
	query := fmt.Sprintf(`
WITH today_attendance AS (
    SELECT
        COUNT(DISTINCT a.employee_id) FILTER (WHERE a.work_day = CURRENT_DATE) AS come_count,
        COUNT(DISTINCT u.employee_id) AS total_count,
        COUNT(u.employee_id) FILTER (WHERE a.employee_id IS NULL) AS absent_count
    FROM users u
    LEFT JOIN attendance a ON a.employee_id = u.employee_id AND a.work_day = CURRENT_DATE
    WHERE u.status = 'false' AND u.deleted_at IS NULL
)
SELECT
    COALESCE(ROUND(100.0 * come_count / GREATEST(1, total_count), 2), 0) AS come_percentage,
    COALESCE(ROUND(100.0 * absent_count / GREATEST(1, total_count), 2), 0) AS absent_percentage
FROM today_attendance;

`)
	//

	var detail PieChartResponse

	row := r.QueryRowContext(ctx, query)
	var comePercentage, absentPercentage float64
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
	query := `
    WITH today_attendance AS (
    SELECT
    COUNT(DISTINCT a.employee_id) FILTER (WHERE a.work_day = CURRENT_DATE) AS come_count,
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

	rows, err := r.QueryContext(ctx, query)
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
            COUNT(DISTINCT a.employee_id) FILTER (WHERE a.work_day = CURRENT_DATE) AS come_count,
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
