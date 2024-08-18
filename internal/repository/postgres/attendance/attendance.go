package attendance

import (
	"context"
	"database/sql"
	"encoding/json"
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

	whereQuery := fmt.Sprintf(`
			WHERE 
				a.deleted_at IS NULL
			`)

	if filter.Search != nil {
		search := strings.Replace(*filter.Search, " ", "", -1)
		search = strings.Replace(search, "'", "''", -1)

		whereQuery += fmt.Sprintf(` AND
		u.employee_id ilike '%s' OR u.full_name ilike '%s'`, "%"+search+"%", "%"+search+"%")
	}
	if filter.DepartmentID != nil {
		whereQuery += fmt.Sprintf(` AND u.department_id = %d`, *filter.DepartmentID)
	}
	if filter.PositionID != nil {
		whereQuery += fmt.Sprintf(` AND u.position_id = %d`, *filter.PositionID)
	}
	if filter.Status != nil {
		var statusValue string
		if *filter.Status {
			statusValue = "true"
		} else {
			statusValue = "false"
		}
		whereQuery += fmt.Sprintf(" AND a.status = %s", statusValue)
	}

	if filter.Date != nil {
		Date, err := time.Parse("2006-01-02", *filter.Date)
		if err != nil {
			return []GetListResponse{}, 0, web.NewRequestError(errors.Wrap(err, "date parse"), http.StatusBadRequest)
		}
		date := Date.Format("2006-01-02")

		whereQuery += fmt.Sprintf(" AND a.work_day = '%s'", date)
	} else {
		today := time.Now().Format("2006-01-02")
		whereQuery += fmt.Sprintf(" AND a.work_day = '%s'", today)
	}
	orderQuery := "ORDER BY a.created_at desc"
	fmt.Println("DAte:", filter.Date)
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
	 FROM   attendance as a
		LEFT JOIN users u ON a.employee_id=u.employee_id
		LEFT JOIN department d ON u.department_id=d.id	
		LEFT JOIN position p ON u.position_id=p.id

		%s %s %s %s
	`, whereQuery, orderQuery, limitQuery, offsetQuery)

	rows, err := r.QueryContext(ctx, query)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, 0, web.NewRequestError(postgres.ErrNotFound, http.StatusNotFound)
	}
	if err != nil {
		return nil, 0, web.NewRequestError(errors.Wrap(err, "selecting attendance"), http.StatusInternalServerError)
	}
	var list []GetListResponse

	for rows.Next() {
		var detail GetListResponse
		var workDayString string
		var leaveTimeBytes []byte
		var comeTimeBytes []byte

		if err = rows.Scan(
			&detail.EmployeeID,
			&detail.Fullname,
			&detail.DepartmentID,
			&detail.Department,
			&detail.PositionID,
			&detail.Position,
			&workDayString,
			&detail.Status,
			&comeTimeBytes,
			&leaveTimeBytes); err != nil {
			return nil, 0, web.NewRequestError(errors.Wrap(err, "scanning attendance list"), http.StatusBadRequest)
		}

		workDay, err := date.ParseDate(workDayString)
		if err != nil {
			return nil, 0, web.NewRequestError(errors.Wrap(err, "converting work_day to date.Date"), http.StatusBadRequest)
		}
		detail.WorkDay = &workDay // Convert the string to date.Date

		// Convert the byte array to time.Time
		comeTime, err := time.Parse("15:04:05", string(comeTimeBytes))
		if err != nil {
			return nil, 0, web.NewRequestError(errors.Wrap(err, "converting come_time to time.Time"), http.StatusBadRequest)
		}
		detail.ComeTime = &comeTime

		if leaveTimeBytes != nil {
			leaveTime, err := time.Parse("15:04:05", string(leaveTimeBytes))
			if err != nil {
				return nil, 0, web.NewRequestError(errors.Wrap(err, "converting leave_time to time.Time"), http.StatusBadRequest)
			}
			detail.LeaveTime = &leaveTime

			// Calculate the time difference between leave_time and come_time
			timeDiff := leaveTime.Sub(comeTime)

			// Calculate total hours and minutes
			hours := int(timeDiff.Hours())
			minutes := int(timeDiff.Minutes()) % 60

			// Format the total hours as HH:MM
			totalHours := fmt.Sprintf("%02d:%02d", hours, minutes)

			detail.TotalHours = totalHours
		} else {
			detail.TotalHours = ""
		}

		list = append(list, detail)
	}

	countQuery := fmt.Sprintf(`
		SELECT
			count(a.id)
		FROM
		    attendance as a
		LEFT JOIN users u ON a.employee_id=u.employee_id
		LEFT JOIN department d ON u.department_id=d.id	
		LEFT JOIN position p ON u.position_id=p.id
		%s
	`, whereQuery)

	countRows, err := r.QueryContext(ctx, countQuery)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, 0, web.NewRequestError(postgres.ErrNotFound, http.StatusBadRequest)
	}
	if err != nil {
		return nil, 0, web.NewRequestError(errors.Wrap(err, "selecting attendance"), http.StatusInternalServerError)
	}

	count := 0

	for countRows.Next() {
		if err = countRows.Scan(&count); err != nil {
			return nil, 0, web.NewRequestError(errors.Wrap(err, "scanning attendance count"), http.StatusInternalServerError)
		}
	}

	return list, count, nil
}

func (r *GetListResponse) MarshalJSON() ([]byte, error) {
	type Alias GetListResponse
	aux := &struct {
		ComeTime  string `json:"come_time,omitempty"`
		LeaveTime string `json:"leave_time,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}

	if r.ComeTime != nil {
		aux.ComeTime = r.ComeTime.Format("15:04")
	}

	if r.LeaveTime != nil {
		aux.LeaveTime = r.LeaveTime.Format("15:04")
	}

	return json.Marshal(aux)
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
			a.leave_time,
			(a.leave_time-a.come_time) as total_hours
		FROM   attendance as a
		LEFT JOIN users u ON a.employee_id=u.employee_id
		LEFT JOIN department d ON u.department_id=d.id	
		LEFT JOIN position p ON u.position_id=p.id
		WHERE  a.deleted_at is NULL and a.id= %d
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
		&detail.TotalHours,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return GetDetailByIdResponse{}, web.NewRequestError(postgres.ErrNotFound, http.StatusBadRequest)
	}

	workDay, err := date.ParseDate(workDayString)
	if err != nil {
		return GetDetailByIdResponse{}, web.NewRequestError(errors.Wrap(err, "parsing come_time"), http.StatusBadRequest)
	}
	detail.WorkDay = &workDay
	// Convert the byte array to time.Time
	comeTime, err := time.Parse("15:04:05", string(comeTimeBytes))
	if err != nil {
		return GetDetailByIdResponse{}, web.NewRequestError(errors.Wrap(err, "converting come_time to time.Time"), http.StatusBadRequest)
	}
	detail.ComeTime = &comeTime

	if leaveTimeBytes != nil {
		leaveTime, err := time.Parse("15:04:05", string(leaveTimeBytes))
		if err != nil {
			return GetDetailByIdResponse{}, web.NewRequestError(errors.Wrap(err, "converting leave_time to time.Time"), http.StatusBadRequest)
		}
		detail.LeaveTime = &leaveTime

		// Calculate the time difference between leave_time and come_time
		timeDiff := leaveTime.Sub(comeTime)

		// Calculate total hours and minutes
		hours := int(timeDiff.Hours())
		minutes := int(timeDiff.Minutes()) % 60

		// Format the total hours as HH:MM
		totalHours := fmt.Sprintf("%02d:%02d", hours, minutes)

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

	response.EmployeeID = request.EmployeeID
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

// func (r Repository) CreateByPhone(ctx context.Context, request EnterRequest) (CreateResponse, error) {
// 	claims, err := r.CheckClaims(ctx)
// 	if err != nil {
// 		return CreateResponse{}, err
// 	}

// 	if err := r.ValidateStruct(&request, "Latitude", "Longitude"); err != nil {
// 		return CreateResponse{}, err
// 	}

// 	var response CreateResponse

// 	response.EmployeeID = request.EmployeeID
// 	response.ComeTime = time.Now().Format("15:04")
// 	response.WorkDay = time.Now().Format("2006-01-02")
// 	response.CreatedAt = time.Now()
// 	response.CreatedBy = claims.UserId

// 	_, err = r.NewInsert().Model(&response).Returning("id").Exec(ctx, &response.ID)
// 	if err != nil {
// 		return CreateResponse{}, web.NewRequestError(errors.Wrap(err, "creating attendance by phone"), http.StatusBadRequest)
// 	}

//		return response, nil
//	}
func (r Repository) CreateByPhone(ctx context.Context, request EnterRequest) (CreateResponse, error) {
	claims, err := r.CheckClaims(ctx)
	if err != nil {
		return CreateResponse{}, err
	}

	if err := r.ValidateStruct(&request, "Latitude", "Longitude"); err != nil {
		return CreateResponse{}, err
	}

	var response CreateResponse

	response.EmployeeID = request.EmployeeID
	response.ComeTime = time.Now().Format("15:04")
	response.WorkDay = time.Now().Format("2006-01-02")
	response.CreatedAt = time.Now()
	response.CreatedBy = claims.UserId

	// query := `
	//      Select
	// `
	// Create attendance entry
	_, err = r.NewInsert().Model(&response).Returning("id").Exec(ctx, &response.ID)
	if err != nil {
		return CreateResponse{}, web.NewRequestError(errors.Wrap(err, "creating attendance by phone"), http.StatusBadRequest)
	}

	// Update user's status to true
	q := r.NewUpdate().Table("users").Where("deleted_at IS NULL AND employee_id = ?", *request.EmployeeID).Set("status = true")

	_, err = q.Exec(ctx)
	if err != nil {
		return CreateResponse{}, web.NewRequestError(errors.Wrap(err, "updating user's status by phone when enter "), http.StatusBadRequest)
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

	comeTime, err := time.Parse("15:04", request.ComeTime)
	if err != nil {
		return web.NewRequestError(errors.Wrap(err, "parsing come time"), http.StatusBadRequest)
	}

	var leaveTime *time.Time
	if request.LeaveTime != "" {
		t, err := time.Parse("15:04", request.LeaveTime)
		if err != nil {
			return web.NewRequestError(errors.Wrap(err, "parsing leave time"), http.StatusBadRequest)
		}
		leaveTime = &t
	}

	q := r.NewUpdate().Table("attendance").Where("deleted_at IS NULL AND id = ?", request.ID)
	q.Set("come_time=?", comeTime)
	q.Set("leave_time=?", leaveTime)
	q.Set("work_day=?", request.WorkDay)
	q.Set("status=?", request.Status)
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
	if request.Status != nil {
		q.Set("status = ?", request.Status)
	}
	q.Set("updated_at = ?", time.Now())
	q.Set("updated_by = ?", claims.UserId)

	_, err = q.Exec(ctx)
	if err != nil {
		return web.NewRequestError(errors.Wrap(err, "updating attendance"), http.StatusBadRequest)
	}

	return nil
}

func (r Repository) ExitByPhone(ctx context.Context, request ExitByPhoneRequest) (ExitResponse, error) {
	claims, err := r.CheckClaims(ctx)
	if err != nil {
		return ExitResponse{}, err
	}

	if err := r.ValidateStruct(&request, "Latitude", "Longitude"); err != nil {
		return ExitResponse{}, err
	}

	var response ExitResponse

	response.EmployeeID = request.EmployeeID
	response.LeaveTime = time.Now().Format("15:04")
	response.CreatedAt = time.Now()
	response.CreatedBy = claims.UserId

	// Update the attendance record
	q := r.NewUpdate().Table("attendance").Where("deleted_at IS NULL AND employee_id = ? AND status = true", request.EmployeeID)
	q.Set("leave_time", time.Now().Format("15:04"))
	q.Set("status", false)
	q.Set("updated_at", time.Now())
	q.Set("updated_by", claims.UserId)

	_, err = q.Exec(ctx)
	if err != nil {
		return ExitResponse{}, web.NewRequestError(errors.Wrap(err, "updating attendance by phone"), http.StatusBadRequest)
	}

	// Update the user's status to false
	userUpdate := r.NewUpdate().Table("users").Where("deleted_at IS NULL AND employee_id = ? AND status = true", request.EmployeeID)
	userUpdate.Set("status", false)

	_, err = userUpdate.Exec(ctx)
	if err != nil {
		return ExitResponse{}, web.NewRequestError(errors.Wrap(err, "updating user's status by phone when exit"), http.StatusBadRequest)
	}

	return response, nil
}

func (r Repository) Delete(ctx context.Context, id int) error {
	return r.DeleteRow(ctx, "attendance", id)
}

func (r Repository) GetStatistics(ctx context.Context) (GetStatisticResponse, error) {
	var response GetStatisticResponse

	query := `
	SELECT
		(SELECT COUNT(DISTINCT employee_id) FROM users) AS total_employee,
		(SELECT COUNT(employee_id) FROM attendance WHERE come_time >= '09:00' AND come_time < '10:00') AS on_time,
		(SELECT COUNT(employee_id) FROM users WHERE status = 'false') AS absent,
		(SELECT COUNT(employee_id) FROM attendance WHERE come_time >= '10:00') AS late_arrival,
		(SELECT COUNT(employee_id) FROM attendance WHERE leave_time < '18:00') AS early_departures,
		(SELECT COUNT(employee_id) FROM attendance WHERE come_time < '09:00') AS time_off
	`

	err := r.DB.QueryRowContext(ctx, query).Scan(
		&response.TotalEmployee,
		&response.OnTime,
		&response.Absent,
		&response.LateArrival,
		&response.EarlyDepartures,
		&response.TimeOff,
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
            COUNT(a.status = true ) AS come_count,
            COUNT( u.status  = false ) AS absent_count,
            COUNT(u.employee_id) AS total_count
        FROM attendance a
        JOIN users u ON a.employee_id = u.employee_id
        WHERE a.work_day = CURRENT_DATE
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
            COUNT(CASE WHEN a.status =  true THEN 1 END) AS come_count,
            COUNT(u.employee_id) AS total_count,
            u.department_id,
            d.name
        FROM attendance a
        JOIN users u ON a.employee_id = u.employee_id
        JOIN department d ON d.id = u.department_id
        GROUP BY u.department_id, d.name
    )
    SELECT
        d.name AS department,
        COALESCE(ROUND(100.0 * come_count / GREATEST(1, total_count), 2), 0) AS percentage
    FROM today_attendance
    JOIN department d ON d.id = today_attendance.department_id;
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

	// Calculate start and end dates directly without string conversion
	startDate := time.Date(filter.Month.Year(), filter.Month.Month(), startDay, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(filter.Month.Year(), filter.Month.Month(), endDay, 23, 59, 59, 999999999, time.UTC)

	// Use parameterized query to avoid SQL injection vulnerabilities
	query := `
	WITH today_attendance AS (
		SELECT
			a.work_day,
			COUNT(CASE WHEN a.status = true THEN 1 END) AS come_count,
			COUNT(u.employee_id) AS total_count
		FROM attendance a
		JOIN users u ON a.employee_id = u.employee_id
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
		return []GraphResponse{}, web.NewRequestError(errors.Wrap(err, "selecting attendance query "), http.StatusInternalServerError)
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, startDate, endDate)
	if err != nil {
		return []GraphResponse{}, web.NewRequestError(errors.Wrap(err, "selecting attendance filter"), http.StatusInternalServerError)
	}
	defer rows.Close()

	var list []GraphResponse

	for rows.Next() {
		var detail GraphResponse
		var workDayString string // Scan the work_day as a string

		if err = rows.Scan(&workDayString, &detail.Percentage); err != nil {
			return []GraphResponse{}, web.NewRequestError(errors.Wrap(err, "scanning Graph response"), http.StatusBadRequest)
		}

		workDay, err := date.ParseDate(workDayString) // Parse the string after scanning
		if err != nil {
			return []GraphResponse{}, web.NewRequestError(errors.Wrap(err, "converting work_day to date.Date"), http.StatusBadRequest)
		}
		detail.WorkDay = &workDay
		list = append(list, detail)
	}

	return list, nil
}
