package user

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
	"university-backend/foundation/web"
	"university-backend/internal/auth"
	"university-backend/internal/entity"
	"university-backend/internal/pkg/repository/postgresql"
	"university-backend/internal/repository/postgres"
	"university-backend/internal/service/hashing"

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
	if filter.DepartmentID != nil {
		whereQuery += fmt.Sprintf(` AND u.department_id = %d`, *filter.DepartmentID)
	}
	if filter.PositionID != nil {
		whereQuery += fmt.Sprintf(` AND u.position_id = %d`, *filter.PositionID)
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

// Create creates new users from an Excel file.
func (r Repository) Create(ctx context.Context, request ExcellRequest) (int, error) {
	claims, err := r.CheckClaims(ctx, auth.RoleAdmin)
	if err != nil {
		return 0, err
	}

	// Validate the ExcellRequest struct
	if err := r.ValidateStruct(&request); err != nil {
		return 0, err
	}
	file, err := request.Excell.Open()
	if err != nil {
		return 0, web.NewRequestError(errors.Wrap(err, "opening excel file"), http.StatusBadRequest)
	}
	defer file.Close()

	// Create a temporary file to store the uploaded Excel file
	tempFile, err := ioutil.TempFile("", "excel-*.xlsx")
	if err != nil {
		return 0, web.NewRequestError(errors.Wrap(err, "creating temporary file"), http.StatusInternalServerError)
	}
	defer tempFile.Close()
	defer os.Remove(tempFile.Name())

	// Copy the uploaded Excel file to the temporary file
	_, err = io.Copy(tempFile, file) // Use the opened file as the reader
	if err != nil {
		return 0, web.NewRequestError(errors.Wrap(err, "copying excel file"), http.StatusInternalServerError)
	}

	// Read the Excel file and parse the data
	excelData, err := hashing.ExcelReader(tempFile.Name())
	if err != nil {
		return 0, web.NewRequestError(errors.Wrap(err, "reading excel data"), http.StatusBadRequest)
	}

	// Check if there are any data rows
	if len(excelData) == 0 {
		return 0, web.NewRequestError(errors.New("no data found in Excel file"), http.StatusBadRequest)
	}

	// Start a transaction
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, web.NewRequestError(errors.Wrap(err, "starting transaction"), http.StatusInternalServerError)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback() // Rollback the transaction if an error occurred
		} else {
			_ = tx.Commit()
		}
	}()

	// Create a new user based on the data from the Excel file
	createdCount := 0 // Initialize the count
	for _, data := range excelData {
		// Validate the data from the Excel file
		if err := r.ValidateStruct(&data, "employee_id", "password", "full_name"); err != nil {
			return 0, err
		}

		// Check for duplicate employee_id
		var existingEmployee CreateResponse
		err := tx.NewSelect().Model(&existingEmployee).Where("employee_id = ?", data.EmployeeID).Scan(ctx)
		if err == nil { // Employee with the same ID exists
			fmt.Println("Error: Duplicate employee_id:", data.EmployeeID)
			return 0, web.NewRequestError(errors.New("Duplicate employee_id found"), http.StatusBadRequest) // Return an error
		}

		// Hash the password
		hash, err := bcrypt.GenerateFromPassword([]byte(data.Password), bcrypt.DefaultCost)
		if err != nil {
			return 0, web.NewRequestError(errors.Wrap(err, "hashing password"), http.StatusInternalServerError)
		}
		hashedPassword := string(hash)

		if data.Role == "" {
			// Option 1: Set a default value
			data.Role = "EMPLOYEE" // Replace with your default role
			// Option 2: Return an error
			return 0, web.NewRequestError(errors.New("Role cannot be empty"), http.StatusBadRequest)
		}
		fmt.Println("Data:da", data)
		// Create a new user object
		user := CreateResponse{
			EmployeeID:   &data.EmployeeID,
			Password:     &hashedPassword,
			Role:         data.Role,
			FullName:     &data.FullName,
			DepartmentID: &data.DepartmentID,
			PositionID:   &data.PositionID,
			Phone:        &data.Phone,
			Email:        &data.Email,
			CreatedAt:    time.Now(),
			CreatedBy:    claims.UserId,
		}

		// Validate the user object
		if err := r.ValidateStruct(&user); err != nil {
			return 0, err
		}

		// Save the new user to the database within the transaction
		_, err = tx.NewInsert().Model(&user).Returning("id").Exec(ctx, &user.ID)
		if err != nil {
			// If an error occurs during insertion, rollback the transaction
			fmt.Printf("Error inserting employee: %v, Error: %v\n", data.EmployeeID, err) // Log the error and employee ID
			return 0, web.NewRequestError(errors.Wrapf(err, "error inserting employee %s", data.EmployeeID), http.StatusBadRequest)
		}

		createdCount++
	}

	return createdCount, err
}

// func (r Repository) Create(ctx context.Context, request ExcellRequest) (CreateResponse, error) {
// 	claims, err := r.CheckClaims(ctx, auth.RoleAdmin)
// 	if err != nil {
// 		return CreateResponse{}, err
// 	}
// 	if err := r.ValidateStruct(&request); err != nil {
// 		return CreateResponse{}, err
// 	}

// 	if err := r.ValidateStruct(&request, "EmployeeID", "Password", "FullName"); err != nil {
// 		return CreateResponse{}, err
// 	}
// 	rand.Seed(time.Now().UnixNano())

// 	userIdStatus := true
// 	if err := r.QueryRowContext(ctx,
// 		fmt.Sprintf(`SELECT
//     						CASE WHEN
//     						(SELECT id FROM users WHERE employee_id = '%s' AND deleted_at IS NULL) IS NOT NULL
//     						THEN true ELSE false END`, *request.EmployeeID)).Scan(&userIdStatus); err != nil {
// 		return CreateResponse{}, web.NewRequestError(errors.Wrap(err, "employee_id check"), http.StatusInternalServerError)
// 	}
// 	if userIdStatus {
// 		return CreateResponse{}, web.NewRequestError(errors.Wrap(errors.New(""), "employee_id is used"), http.StatusBadRequest)
// 	}

// 	hash, err := bcrypt.GenerateFromPassword([]byte(*request.Password), bcrypt.DefaultCost)
// 	if err != nil {
// 		return CreateResponse{}, web.NewRequestError(errors.Wrap(err, "hashing password"), http.StatusInternalServerError)
// 	}
// 	hashedPassword := string(hash)

// 	// var response CreateResponse
// 	// role := strings.ToUpper(*request.Role)
// 	// if (role != "EMPLOYEE") && (role != "ADMIN") {
// 	// 	return CreateResponse{}, web.NewRequestError(errors.New("incorrect role. role should be EMPLOYEE or ADMIN"), http.StatusBadRequest)
// 	// }

// 	// response.Role = &role
// 	// response.FullName = request.FullName
// 	// response.EmployeeID = request.EmployeeID
// 	// response.Password = &hashedPassword
// 	// response.DepartmentID = request.DepartmentID
// 	// response.PositionID = request.PositionID
// 	// response.Phone = request.Phone
// 	// response.Email = request.Email
// 	// response.CreatedAt = time.Now()
// 	// response.CreatedBy = claims.UserId

// 	// Start transaction
// 	tx, err := r.DB.Begin()
// 	if err != nil {
// 		return CreateResponse{}, web.NewRequestError(errors.Wrap(err, "starting transaction"), http.StatusInternalServerError)
// 	}
// 	defer func() {
// 		if err != nil {
// 			tx.Rollback()
// 		}
// 	}()
// 	_, err = r.NewInsert().Model(&response).Returning("id").Exec(ctx, &response.ID)
// 	if err != nil {
// 		tx.Rollback()
// 		return CreateResponse{}, web.NewRequestError(errors.Wrap(err, "creating user"), http.StatusBadRequest)
// 	}
// 	var excelData []hashing.UserExcellData

// 	var fileUrl string
// 	if request.Excell != nil {
// 		fileUrl, _, _, err = commands.Upload(ctx, request.Excell, "user-create/excel")
// 		if err != nil {
// 			tx.Rollback()
// 			return CreateResponse{}, web.NewRequestError(errors.Wrap(err, "upload excel"), http.StatusInternalServerError)
// 		}
// 	}

// 	dir, err := os.Getwd()
// 	if err != nil {
// 		tx.Rollback()
// 		return CreateResponse{}, web.NewRequestError(errors.Wrap(err, "getting uploaded excel path"), http.StatusInternalServerError)
// 	}

// 	if fileUrl != "" {

// 		excelData, err = hashing.ExcelReader(dir + fileUrl)
// 		if err != nil {
// 			tx.Rollback()
// 			return CreateResponse{}, web.NewRequestError(errors.Wrap(err, "reading excel data"), http.StatusInternalServerError)
// 		}

// 		for _, data := range excelData {
// 			var user entity.User

// 			err := r.NewSelect().Model(&user).Where("employee_id = ?", data.EmployeeID).Scan(ctx)
// 			if err != nil {
// 				tx.Rollback()
// 				return CreateResponse{}, web.NewRequestError(errors.Wrap(err, "getting user detail by code"), http.StatusBadRequest)
// 			}

// 			q := r.NewUpdate().Table("users").Where("deleted_at IS NULL AND employee_id = ? ",  data.EmployeeID)

// 			q.Set("updated_at = ?", time.Now())
// 			q.Set("updated_by = ?", claims.UserId)

// 			_, err = q.Exec(ctx)
// 			if err != nil {
// 				tx.Rollback()
// 				return CreateResponse{}, web.NewRequestError(errors.Wrap(err, "updating user"), http.StatusBadRequest)
// 			}

// 		}
// 	}
// 	// Commit transaction
// 	if err = tx.Commit(); err != nil {
// 		tx.Rollback()
// 		return CreateResponse{}, web.NewRequestError(errors.Wrap(err, "committing transaction"), http.StatusInternalServerError)
// 	}

// 	return response, nil
// }

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

func (r Repository) GetMonthlyStatistics(ctx context.Context, request MonthlyStatisticRequest) (MonthlyStatisticResponse, error) {
	claims, err := r.CheckClaims(ctx)
	if err != nil {
		return MonthlyStatisticResponse{}, web.NewRequestError(errors.New("error getting user by token"), http.StatusBadRequest)
	}

	// Calculate the start and end dates of the month
	startDate := request.Month
	endDate := startDate.AddDate(0, 1, -1) // Last day of the month

	// Convert dates to strings in the format 'YYYY-MM-DD'
	startDateStr := startDate.Format("2006-01-02")
	endDateStr := endDate.Format("2006-01-02")

	// Initialize the response with 0 values
	list := MonthlyStatisticResponse{
		EarlyCome:  new(int),
		EarlyLeave: new(int),
		Absent:     new(int),
		Late:       new(int),
	}

	// Query for monthly statistics
	monthlyQuery := `
		SELECT
			COALESCE(SUM(CASE WHEN a.come_time <= '09:00' THEN 1 ELSE 0 END), 0) AS early_come,
			COALESCE(SUM(CASE WHEN a.leave_time < '18:00' THEN 1 ELSE 0 END), 0) AS early_leave,
			COALESCE(SUM(CASE WHEN u.status = 'false' THEN 1 ELSE 0 END), 0) AS absent,
			COALESCE(SUM(CASE WHEN a.come_time >= '10:00' THEN 1 ELSE 0 END), 0) AS late
		FROM attendance a
		JOIN users u ON a.employee_id = u.employee_id
		WHERE a.deleted_at IS NULL
		AND u.id = $1
		AND a.work_day BETWEEN $2 AND $3;
	`

	monthlyStmt, err := r.Prepare(monthlyQuery)
	if err != nil {
		return MonthlyStatisticResponse{}, web.NewRequestError(errors.Wrap(err, "preparing monthly query"), http.StatusInternalServerError)
	}
	defer monthlyStmt.Close()

	err = monthlyStmt.QueryRowContext(ctx, claims.UserId, startDateStr, endDateStr).Scan(
		list.EarlyCome,
		list.EarlyLeave,
		list.Absent,
		list.Late,
	)
	if err != nil {
		return MonthlyStatisticResponse{}, web.NewRequestError(errors.Wrap(err, "scanning monthly statistics"), http.StatusInternalServerError)
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

	// Generate all dates within the interval
	var allDates []time.Time
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		allDates = append(allDates, d)
	}

	// Query for interval data
	intervalQuery := `
		SELECT
			a.work_day,
			COALESCE(TO_CHAR(a.come_time, 'HH24:MI'), '00:00') AS come_time,
			COALESCE(TO_CHAR(a.leave_time, 'HH24:MI'), '00:00') AS leave_time,
			COALESCE(SUM(EXTRACT(EPOCH FROM (ap.leave_time - ap.come_time)) / 60), 0) AS total_minutes
		FROM attendance a
		JOIN users u ON a.employee_id = u.employee_id
		LEFT JOIN attendance_period ap ON a.id = ap.attendance_id
		WHERE a.deleted_at IS NULL
			AND u.id = $1
			AND a.work_day BETWEEN $2 AND $3
		GROUP BY a.work_day, a.come_time, a.leave_time
		ORDER BY a.work_day;
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

	// Map to store retrieved data by date
	dataMap := make(map[string]StatisticResponse)
	for rows.Next() {
		var detail StatisticResponse
		var totalMinutes float64
		err := rows.Scan(
			&detail.WorkDay,
			&detail.ComeTime,
			&detail.LeaveTime,
			&totalMinutes,
		)
		if err != nil {
			return nil, web.NewRequestError(errors.Wrap(err, "scanning interval statistics"), http.StatusInternalServerError)
		}

		// Convert total minutes to HH:MM
		hours := int(totalMinutes) / 60
		minutes := int(totalMinutes) % 60
		detail.TotalHours = fmt.Sprintf("%02d:%02d", hours, minutes)
		dataMap[*detail.WorkDay] = detail
	}

	// Generate the final list of responses, filling in missing dates with default values
	var list []StatisticResponse
	for _, date := range allDates {
		dateStr := date.Format("2006-01-02")
		if data, found := dataMap[dateStr]; found {
			list = append(list, data)
		} else {
			list = append(list, StatisticResponse{
				WorkDay:    &dateStr,
				ComeTime:   ptr("00:00"),
				LeaveTime:  ptr("00:00"),
				TotalHours: "00:00",
			})
		}
	}

	return list, nil
}

// Utility function to return a pointer to a string
func ptr(s string) *string {
	return &s
}

func (r Repository) GetEmployeeDashboard(ctx context.Context) (DashboardResponse, error) {
	claims, err := r.CheckClaims(ctx)
	if err != nil {
		return DashboardResponse{}, err
	}
	workDay := time.Now().Format("2006-01-02")

	var detail DashboardResponse
	var totalMinutes int
	query := fmt.Sprintf(`
        SELECT
    MAX(ap.come_time) AS come_time,  -- Use MAX to get the latest come_time
    MAX(a.leave_time) AS leave_time, -- Use MAX to get the latest leave_time
    COALESCE(SUM(EXTRACT(EPOCH FROM (ap.leave_time - ap.come_time))/ 60)::INT, 0) AS total_hours
FROM attendance AS a
JOIN users AS u ON u.employee_id = a.employee_id
JOIN attendance_period AS ap ON ap.attendance_id = a.id
WHERE a.work_day= '%s'
AND ap.work_day= '%s'
AND a.deleted_at IS NULL
AND u.deleted_at IS NULL
AND u.id = %d
GROUP BY a.employee_id
ORDER BY come_time DESC
LIMIT 1;            
	`, workDay,workDay,claims.UserId)
	err = r.QueryRowContext(ctx, query).Scan(
		&detail.ComeTime,
		&detail.LeaveTime,
		&totalMinutes,
	)

	hours := totalMinutes / 60
	minutes := totalMinutes % 60
	totalHours := fmt.Sprintf("%02d:%02d", hours, minutes)
	detail.TotalHours = totalHours

	if errors.Is(err, sql.ErrNoRows) {
		return DashboardResponse{}, nil
	}
	if err != nil {
		return DashboardResponse{}, web.NewRequestError(errors.Wrap(err, "selecting user detail on dashboard"), http.StatusBadRequest)
	}
	return detail, nil
}
