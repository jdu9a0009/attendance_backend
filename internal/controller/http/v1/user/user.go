package user

import (
	"attendance/backend/foundation/web"
	"attendance/backend/internal/pkg/config"
	"attendance/backend/internal/repository/postgres/user"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"

	"github.com/Azure/go-autorest/autorest/date"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v4"
	_ "github.com/lib/pq" // PostgreSQL driver
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
func (uc Controller) ExportEmployee(c *web.Context) error {
	// Generate the Excel file containing employee data
	xlsxFilename, err := uc.user.ExportEmployee(c.Ctx)
	if err != nil {
		return c.RespondError(err) // Handle any error from generating the Excel file
	}

	file, err := os.Open(xlsxFilename)
	if err != nil {
		return c.RespondError(err)
	}
	defer file.Close()

	// Set the content type for Excel files
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=\"employee_list.xlsx\"")

	_, err = io.Copy(c.Writer, file)
	if err != nil {
		return c.RespondError(err)
	}
	os.Remove("employee_list.xlsx")
	return nil
}
func (uc Controller) ExportTemplate(c *web.Context) error {
	// Generate the Excel file containing employee data
	xlsxFilename, err := uc.user.ExportTemplate(c.Ctx)
	if err != nil {
		return c.RespondError(err) // Handle any error from generating the Excel file
	}

	file, err := os.Open(xlsxFilename)
	if err != nil {
		return c.RespondError(err)
	}
	defer file.Close()

	// Set the content type for Excel files
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=\"template.xlsx\"")

	_, err = io.Copy(c.Writer, file)
	if err != nil {
		return c.RespondError(err)
	}

	// os.Remove("template.xls")
	return nil
}

func (uc Controller) CreateUser(c *web.Context) error {
	var request user.CreateRequest
	if err := c.BindFunc(&request); err != nil {
		return c.RespondError(err)
	}
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

	var response int
	var incomplete []int // Declare outside switch to use later
	var err error        // Declare error variable

	switch request.Mode {
	case 1: // Create mode
		response, incomplete, err = uc.user.CreateByExcell(c.Ctx, request)
	case 2: // Update mode
		response, incomplete, err = uc.user.UpdateByExcell(c.Ctx, request)
	case 3: // Delete mode
		response, incomplete, err = uc.user.DeleteByExcell(c.Ctx, request)
	default:
		return c.RespondError(errors.New("invalid mode specified"))
	}

	// Check for any error that occurred during the operation
	if err != nil {
		return c.RespondError(err)
	}

	return c.Respond(map[string]interface{}{
		"Successfully Users":           response,
		"Incomplete  Users Excell Row": incomplete,
		"status":                       true,
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
		"data":     response,
		"status":   true,
	}, http.StatusOK)
}

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins
		return true
	},
}

// Handle WebSocket connection for real-time updates

// Create a function to connect to the database using pgx
func ConnectDB(ctx context.Context, dsn string) (*pgx.Conn, error) {
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}
	return conn, nil
}

// Update your waitForNotification function
func waitForNotification(conn *pgx.Conn) (string, error) {
	// Wait for the next notification
	notification, err := conn.WaitForNotification(context.Background())
	if err != nil {
		return "", err
	}

	// Return the payload of the notification
	return notification.Payload, nil
}

// Update your WebSocket handler
func (uc Controller) GetDashboardListWS(w http.ResponseWriter, r *http.Request) {
	log.Println("Attempting to upgrade connection to WebSocket...")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket Upgrade Error: %v", err)
		http.Error(w, "Could not open WebSocket connection", http.StatusBadRequest)
		return
	}
	defer conn.Close()
	log.Println("WebSocket connection established")

	ctx := r.Context()
	yamlConfig, err := config.NewConfig()
	if err != nil {
		log.Printf("Error loading configuration: %v", err)
		return
	}

	dsn := fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=disable", yamlConfig.DBUsername, yamlConfig.DBPassword, yamlConfig.DBHost, yamlConfig.DBPort, yamlConfig.DBName)
	dbConn, err := ConnectDB(ctx, dsn)
	if err != nil {
		log.Printf("Database connection error: %v", err)
		conn.WriteJSON(map[string]interface{}{"error": "Failed to connect to the database"})
		return
	}
	defer dbConn.Close(ctx)

	// Start listening for changes on the 'attendance_changes' channel
	_, err = dbConn.Exec(ctx, `LISTEN attendance_changes`)
	if err != nil {
		log.Printf("Failed to start listening for attendance changes: %v", err)
		conn.WriteJSON(map[string]interface{}{"error": "Failed to listen for database changes"})
		return
	}

	// Send initial data
	filter := user.Filter{}
	data, _, err := uc.user.GetDashboardList(ctx, filter)
	if err != nil {
		log.Printf("Error fetching initial dashboard data: %v", err)
		conn.WriteJSON(map[string]interface{}{"error": "Failed to load dashboard data"})
		return
	}
	log.Printf("Initial dashboard data: %v", data)
	conn.WriteJSON(map[string]interface{}{"data": data})

	// Listen for notifications in a loop
	for {
		select {
		case <-ctx.Done():
			log.Println("WebSocket connection closed")
			return
		default:
			notification, err := waitForNotification(dbConn)
			if err != nil {
				log.Printf("Error waiting for notification: %v", err)
				continue
			}

			log.Printf("Received notification: %s", notification)

			// Fetch updated dashboard data only once per notification
			updatedData, _, err := uc.user.GetDashboardList(ctx, filter)
			if err != nil {
				log.Printf("Error fetching updated dashboard data: %v", err)
				continue
			}

			// Send updated data to WebSocket client
			if err := conn.WriteJSON(map[string]interface{}{"data": updatedData}); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}
		}
	}
}
