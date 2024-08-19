package attendance

import (
	"errors"
	"math"
	"net/http"
	"reflect"
	"strconv"
	"university-backend/foundation/web"
	"university-backend/internal/repository/postgres/attendance"

	"github.com/Azure/go-autorest/autorest/date"
)

type Controller struct {
	attendance Attendance
}

func NewController(attendance Attendance) *Controller {
	return &Controller{attendance}
}

const (
	OfficeLatitude  = 40.730610
	OfficeLongitude = -73.935242
	OfficeRadius    = 200 // in meters
)

//Attendance

func (uc Controller) GetList(c *web.Context) error {
	var filter attendance.Filter

	if limit, ok := c.GetQueryFunc(reflect.Int, "limit").(*int); ok {
		filter.Limit = limit
	}
	if offset, ok := c.GetQueryFunc(reflect.Int, "offset").(*int); ok {
		filter.Offset = offset
	}
	if page, ok := c.GetQueryFunc(reflect.Int, "page").(*int); ok {
		filter.Page = page
	}
	if date, ok := c.GetQueryFunc(reflect.String, "date").(*string); ok {
		filter.Date = date
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
	if status, ok := c.GetQueryFunc(reflect.Bool, "status").(*bool); ok {
		filter.Status = status
	}
	if err := c.ValidQuery(); err != nil {
		return c.RespondError(err)
	}
	list, count, err := uc.attendance.GetList(c.Ctx, filter)
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

func (uc Controller) GetDetailById(c *web.Context) error {
	id := c.GetParam(reflect.Int, "id").(int)

	if err := c.ValidParam(); err != nil {
		return c.RespondError(err)
	}

	response, err := uc.attendance.GetDetailById(c.Ctx, id)
	if err != nil {
		return c.RespondError(err)
	}

	return c.Respond(map[string]interface{}{
		"data":   response,
		"status": true,
	}, http.StatusOK)
}

func (uc Controller) GetStatistics(c *web.Context) error {

	if err := c.ValidParam(); err != nil {
		return c.RespondError(err)
	}

	response, err := uc.attendance.GetStatistics(c.Ctx)
	if err != nil {
		return c.RespondError(err)
	}

	return c.Respond(map[string]interface{}{
		"data":   response,
		"status": true,
	}, http.StatusOK)
}

func (uc Controller) GetPieChartStatistics(c *web.Context) error {

	if err := c.ValidParam(); err != nil {
		return c.RespondError(err)
	}

	response, err := uc.attendance.GetPieChartStatistic(c.Ctx)
	if err != nil {
		return c.RespondError(err)
	}

	return c.Respond(map[string]interface{}{
		"data":   response,
		"status": true,
	}, http.StatusOK)
}
func (uc Controller) GetBarChartStatistics(c *web.Context) error {

	if err := c.ValidParam(); err != nil {
		return c.RespondError(err)
	}

	response, err := uc.attendance.GetBarChartStatistic(c.Ctx)
	if err != nil {
		return c.RespondError(err)
	}

	return c.Respond(map[string]interface{}{
		"data":   response,
		"status": true,
	}, http.StatusOK)
}

func (uc Controller) GetGraphStatistic(c *web.Context) error {
	var filter attendance.GraphRequest
	// Get the 'month' query parameter
	monthStr := c.Query("month")
	if monthStr == "" {
		return c.RespondError(web.NewRequestError(errors.New("month parameter is required"), http.StatusBadRequest))
	}

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

	list, err := uc.attendance.GetGraphStatistic(c.Ctx, filter)
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

// func (uc Controller) GetGraphStatistic(c *web.Context) error {
// 	var request attendance.GraphRequest
// 	if err := c.BindFunc(&request, "month", "interval"); err != nil {
// 		return c.RespondError(err)
// 	}
// 	fmt.Println("controllar request:", request)
// 	response, err := uc.attendance.GetGraphStatistic(c.Ctx, request)
// 	if err != nil {
// 		return c.RespondError(err)
// 	}

// 	fmt.Println("controllar response:", response)
// 	return c.Respond(map[string]interface{}{
// 		"data":   response,
// 		"status": true,
// 	}, http.StatusOK)

// }

func (uc Controller) Create(c *web.Context) error {
	var request attendance.CreateRequest
	if err := c.BindFunc(&request, "Name"); err != nil {
		return c.RespondError(err)
	}

	response, err := uc.attendance.Create(c.Ctx, request)
	if err != nil {
		return c.RespondError(err)
	}

	return c.Respond(map[string]interface{}{
		"data":   response,
		"status": true,
	}, http.StatusOK)
}

func (uc Controller) UpdateAll(c *web.Context) error {
	id := c.GetParam(reflect.Int, "id").(int)

	if err := c.ValidParam(); err != nil {
		return c.RespondError(err)
	}

	var request attendance.UpdateRequest

	if err := c.BindFunc(&request, "EmployeeID"); err != nil {
		return c.RespondError(err)
	}

	request.ID = id

	err := uc.attendance.UpdateAll(c.Ctx, request)
	if err != nil {
		return c.RespondError(err)
	}

	return c.Respond(map[string]interface{}{
		"data":   "ok!",
		"status": true,
	}, http.StatusOK)
}

func (uc Controller) UpdateColumns(c *web.Context) error {
	id := c.GetParam(reflect.Int, "id").(int)

	if err := c.ValidParam(); err != nil {
		return c.RespondError(err)
	}

	var request attendance.UpdateRequest

	if err := c.BindFunc(&request); err != nil {
		return c.RespondError(err)
	}

	request.ID = id

	err := uc.attendance.UpdateColumns(c.Ctx, request)
	if err != nil {
		return c.RespondError(err)
	}

	return c.Respond(map[string]interface{}{
		"data":   "ok!",
		"status": true,
	}, http.StatusOK)
}

func (uc Controller) Delete(c *web.Context) error {
	id := c.GetParam(reflect.Int, "id").(int)

	if err := c.ValidParam(); err != nil {
		return c.RespondError(err)
	}

	err := uc.attendance.Delete(c.Ctx, id)
	if err != nil {
		return c.RespondError(err)
	}

	return c.Respond(map[string]interface{}{
		"data":   "ok!",
		"status": true,
	}, http.StatusOK)
}

func (uc Controller) CreateByPhone(c *web.Context) error {
	var request attendance.EnterRequest
	if err := c.BindFunc(&request, "Latitude,Longitude"); err != nil {
		return c.RespondError(err)
	}
	distance := CalculateDistance(request.Latitude, request.Longitude, OfficeLatitude, OfficeLongitude)

	if distance <= OfficeRadius {
		response, err := uc.attendance.CreateByPhone(c.Ctx, request)
		if err != nil {
			return c.RespondError(err)
		}

		return c.Respond(map[string]interface{}{
			"data":   response,
			"status": true,
		}, http.StatusOK)
	}
	return c.RespondError(web.NewRequestError(errors.New("distance from office is greater than office radius"), http.StatusBadRequest))
}
func (uc Controller) ExitByPhone(c *web.Context) error {
	var request attendance.EnterRequest
	if err := c.BindFunc(&request, "Latitude,Longitude"); err != nil {
		return c.RespondError(err)
	}
	distance := CalculateDistance(request.Latitude, request.Longitude, OfficeLatitude, OfficeLongitude)

	if distance <= OfficeRadius {
		response, err := uc.attendance.ExitByPhone(c.Ctx, request)
		if err != nil {
			return c.RespondError(err)
		}

		return c.Respond(map[string]interface{}{
			"data":   response,
			"status": true,
		}, http.StatusOK)
	}
	return c.RespondError(web.NewRequestError(errors.New("distance from office is greater than office radius"), http.StatusBadRequest))
}

func CalculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	// Haversine formula to calculate the great-circle distance between two points
	R := 6371.0 // Earth's radius in kilometers
	φ1 := lat1 * math.Pi / 180.0
	φ2 := lat2 * math.Pi / 180.0
	Δφ := (lat2 - lat1) * math.Pi / 180.0
	Δλ := (lon2 - lon1) * math.Pi / 180.0

	a := math.Sin(Δφ/2)*math.Sin(Δφ/2) + math.Cos(φ1)*math.Cos(φ2)*math.Sin(Δλ/2)*math.Sin(Δλ/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	distance := R * c * 1000 // Distance in meters

	return distance
}
