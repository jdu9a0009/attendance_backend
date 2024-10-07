package user

import (
	"attendance/backend/foundation/web"
	"attendance/backend/internal/repository/postgres/user"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"

	"github.com/Azure/go-autorest/autorest/date"
)

type Controller struct {
	user User
}

func NewController(user User) *Controller {
	return &Controller{user}
}

// user

func (uc Controller) GetUserList(c *web.Context) error {
	var filter user.Filter

	if limit, ok := c.GetQueryFunc(reflect.Int, "limit").(*int); ok {
		filter.Limit = limit
	}
	if offset, ok := c.GetQueryFunc(reflect.Int, "offset").(*int); ok {
		filter.Offset = offset
	}
	if page, ok := c.GetQueryFunc(reflect.Int, "page").(*int); ok {
		filter.Page = page
	}
	if search, ok := c.GetQueryFunc(reflect.String, "search").(*string); ok {
		filter.Search = search
	}
	if departmentId, ok := c.GetQueryFunc(reflect.Int, "department_id").(*int); ok {
		filter.DepartmentID = departmentId
	}
	if positionId, ok := c.GetQueryFunc(reflect.Int, "position_id").(*int); ok {
		filter.PositionID = positionId
	}
	if err := c.ValidQuery(); err != nil {
		return c.RespondError(err)
	}

	list, count, err := uc.user.GetList(c.Ctx, filter)
	if err != nil {
		return c.RespondError(err)
	}

	return c.Respond(map[string]interface{}{
		"data": map[string]interface{}{
			"results": list,
			"count":   count,
		},
		"status": true,
	}, http.StatusOK)
}

func (uc Controller) GetUserDetailById(c *web.Context) error {
	id := c.GetParam(reflect.Int, "id").(int)

	if err := c.ValidParam(); err != nil {
		return c.RespondError(err)
	}

	response, err := uc.user.GetDetailById(c.Ctx, id)
	if err != nil {
		return c.RespondError(err)
	}

	return c.Respond(map[string]interface{}{
		"data":   response,
		"status": true,
	}, http.StatusOK)
}
func (uc Controller) GetQrCodeByEmployeeId(c *web.Context) error {

	// Get the 'employee_id' query parameter
	employeeID := c.Query("employee_id")
	if employeeID == "" {
		return c.RespondError(web.NewRequestError(errors.New("employee_id parameter is required"), http.StatusBadRequest))
	}

	// Call the repository method to get the image file path
	filePath, err := uc.user.GetQrCodeByEmployeeID(c.Ctx, employeeID)
	if err != nil {
		return c.RespondError(err)
	}

	// Open the QR code image file
	file, err := os.Open(filePath)
	if err != nil {
		return c.RespondError(err)
	}
	defer file.Close()

	// Set the content type to PNG
	c.Header("Content-Type", "image/png")
	c.Header("Content-Disposition", "inline; filename="+filepath.Base(filePath))
	// Write the image data to the response
	c.Status(http.StatusOK)
	_, err = io.Copy(c.Writer, file)
	if err != nil {
		return c.RespondError(err)
	}

	return nil
}
func (uc Controller) GetQrCodeList(c *web.Context) error {
	// Generate the PDF containing QR codes for all employees
	pdfFilename, err := uc.user.GetQrCodeList(c.Ctx)
	if err != nil {
		return c.RespondError(err)
	}
	file, err := os.Open(pdfFilename)
	if err != nil {
		return c.RespondError(err)
	}
	defer file.Close()
	// Set the content type to PDF
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=\"qr_employees.pdf\"")
	// Write the PDF to the response
	_, err = io.Copy(c.Writer, file)
	if err != nil {
		return c.RespondError(err)
	}
	return nil
}

func (uc Controller) CreateUser(c *web.Context) error {
	var request user.CreateRequest
	if err := c.BindFunc(&request); err != nil {
		return c.RespondError(err)
	}
	fmt.Println("contro", request)
	response, err := uc.user.Create(c.Ctx, request)
	if err != nil {
		return c.RespondError(err)
	}

	return c.Respond(map[string]interface{}{
		"created_data": response,
		"status":       true,
	}, http.StatusOK)
}
func (uc Controller) CreateUserByExcell(c *web.Context) error {
	var request user.ExcellRequest
	if err := c.BindFunc(&request); err != nil {
		return c.RespondError(err)
	}

	response, err := uc.user.CreateByExcell(c.Ctx, request)
	if err != nil {
		return c.RespondError(err)
	}

	return c.Respond(map[string]interface{}{
		"created_data": response,
		"status":       true,
	}, http.StatusOK)
}

func (uc Controller) UpdateUserAll(c *web.Context) error {
	id := c.GetParam(reflect.Int, "id").(int)

	if err := c.ValidParam(); err != nil {
		return c.RespondError(err)
	}

	var request user.UpdateRequest

	if err := c.BindFunc(&request, "UserID", "FirstName", "Surname"); err != nil {
		return c.RespondError(err)
	}

	request.ID = id

	err := uc.user.UpdateAll(c.Ctx, request)
	if err != nil {
		return c.RespondError(err)
	}

	return c.Respond(map[string]interface{}{
		"data":   "ok!",
		"status": true,
	}, http.StatusOK)
}

func (uc Controller) UpdateUserColumns(c *web.Context) error {
	id := c.GetParam(reflect.Int, "id").(int)

	if err := c.ValidParam(); err != nil {
		return c.RespondError(err)
	}

	var request user.UpdateRequest

	if err := c.BindFunc(&request); err != nil {
		return c.RespondError(err)
	}

	request.ID = id

	err := uc.user.UpdateColumns(c.Ctx, request)
	if err != nil {
		return c.RespondError(err)
	}

	return c.Respond(map[string]interface{}{
		"data":   "ok!",
		"status": true,
	}, http.StatusOK)
}

func (uc Controller) DeleteUser(c *web.Context) error {
	id := c.GetParam(reflect.Int, "id").(int)

	if err := c.ValidParam(); err != nil {
		return c.RespondError(err)
	}

	err := uc.user.Delete(c.Ctx, id)
	if err != nil {
		return c.RespondError(err)
	}

	return c.Respond(map[string]interface{}{
		"data":   "ok!",
		"status": true,
	}, http.StatusOK)
}

func (uc Controller) GetStatistics(c *web.Context) error {
	var filter user.StatisticRequest
	// Get the 'month' query parameter
	monthStr := c.Query("month")
	if monthStr == "" {
		return c.RespondError(web.NewRequestError(errors.New("month parameter is required"), http.StatusBadRequest))
	}
	fmt.Println("Month", monthStr)
	parsedMonth, err := date.ParseDate(monthStr)
	if err != nil {
		return c.RespondError(web.NewRequestError(errors.New("invalid date format"), http.StatusBadRequest))
	}
	filter.Month = parsedMonth

	// Get the 'interval' query parameter
	intervalStr := c.Query("interval")
	if intervalStr == "" {
		return c.RespondError(web.NewRequestError(errors.New("interval parameter is required"), http.StatusBadRequest))
	}

	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		return c.RespondError(web.NewRequestError(errors.New("invalid interval format"), http.StatusBadRequest))
	}
	filter.Interval = interval
	list, err := uc.user.GetStatistics(c.Ctx, filter)
	if err != nil {
		return c.RespondError(err)
	}

	return c.Respond(map[string]interface{}{
		"data": map[string]interface{}{
			"results": list,
		},
		"status": true,
	}, http.StatusOK)
}

func (uc Controller) GetMonthlyStatistics(c *web.Context) error {
	var filter user.MonthlyStatisticRequest
	// Get the 'month' query parameter
	monthStr := c.Query("month")
	if monthStr == "" {
		return c.RespondError(web.NewRequestError(errors.New("month parameter is required"), http.StatusBadRequest))
	}
	fmt.Println("Month", monthStr)
	parsedMonth, err := date.ParseDate(monthStr)
	if err != nil {
		return c.RespondError(web.NewRequestError(errors.New("invalid date format"), http.StatusBadRequest))
	}
	filter.Month = parsedMonth
	list, err := uc.user.GetMonthlyStatistics(c.Ctx, filter)
	if err != nil {
		return c.RespondError(err)
	}
	fmt.Println("Clist", list)
	return c.Respond(map[string]interface{}{
		"data": map[string]interface{}{
			"results": list,
		},
		"status": true,
	}, http.StatusOK)
}

func (uc Controller) GetEmployeeDashboard(c *web.Context) error {

	if err := c.ValidParam(); err != nil {
		return c.RespondError(err)
	}

	response, err := uc.user.GetEmployeeDashboard(c.Ctx)
	if err != nil {
		return c.RespondError(err)
	}
	full_name, err := uc.user.GetFullName(c.Ctx)
	if err != nil {
		return c.RespondError(err)
	}

	return c.Respond(map[string]interface{}{
		"employee": full_name,
		"data":      response,
		"status":    true,
	}, http.StatusOK)
}

func (uc Controller) GetDashboardList(c *web.Context) error {
	var filter user.Filter

	if limit, ok := c.GetQueryFunc(reflect.Int, "limit").(*int); ok {
		filter.Limit = limit
	}
	if offset, ok := c.GetQueryFunc(reflect.Int, "offset").(*int); ok {
		filter.Offset = offset
	}
	if page, ok := c.GetQueryFunc(reflect.Int, "page").(*int); ok {
		filter.Page = page
	}

	if err := c.ValidParam(); err != nil {
		return c.RespondError(err)
	}
	list, count, err := uc.user.GetDashboardList(c.Ctx, filter)
	if err != nil {
		return c.RespondError(err)
	}
	department, err := uc.user.GetDepartmentList(c.Ctx)
	if err != nil {
		return c.RespondError(err)
	}

	return c.Respond(map[string]interface{}{
		"data": map[string]interface{}{
			"department":           department,
			"employee_list":        list,
			"total_employee_count": count,
		},
		"status": true,
	}, http.StatusOK)
}
