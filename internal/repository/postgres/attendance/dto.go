package attendance

import (
	"encoding/json"
	"time"

	"github.com/Azure/go-autorest/autorest/date"
	"github.com/uptrace/bun"
)

type Filter struct {
	Limit        *int
	Offset       *int
	Page         *int
	Search       *string
	DepartmentID *int
	PositionID   *int
	Status       *bool
	Date         *time.Time
}

type GetListResponse struct {
	ID           int        `json:"id"`
	DepartmentID *int       `json:"department_id"`
	Department   *string    `json:"department"`
	PositionID   *int       `json:"position_id"`
	Position     *string    `json:"position"`
	EmployeeID   *string    `json:"employee_id"`
	Fullname     *string    `json:"full_name"`
	Status       *bool      `json:"status"`
	WorkDay      *date.Date `json:"work_day"`
	ComeTime     *time.Time `json:"come_time,omitempty"`
	LeaveTime    *time.Time `json:"leave_time,omitempty"`
	TotalHours   string     `json:"total_hourse"`
}

type GetDetailByIdResponse struct {
	ID           int        `json:"id"`
	DepartmentID *int       `json:"department_id"`
	Department   *string    `json:"department"`
	PositionID   *int       `json:"position_id"`
	Position     *string    `json:"position"`
	EmployeeID   *string    `json:"employee_id"`
	Fullname     *string    `json:"full_name"`
	Status       *bool      `json:"status"`
	WorkDay      *date.Date `json:"work_day"`
	ComeTime     *time.Time `json:"come_time,omitempty"`
	LeaveTime    *time.Time `json:"leave_time,omitempty"`
	TotalHours   string     `json:"total_hours"`
}

type CreateRequest struct {
	EmployeeID *string `json:"employee_id" form:"employee_id"`
}

type CreateResponse struct {
	bun.BaseModel `bun:"table:attendance"`

	ID         int         `json:"id" bun:"-"`
	EmployeeID *string     `json:"employee_id" bun:"employee_id"`
	WorkDay    string      `json:"work_day" bun:"work_day"`
	ComeTime   string      `json:"come_time" bun:"come_time"`
	LeaveTime  *string     `json:"leave_time,omitempty" bun:"leave_time"`
	Periods    []time.Time `json:"-" bun:"periods,type:jsonb"`
	CreatedAt  time.Time   `json:"-"          bun:"created_at"`
	CreatedBy  int         `json:"-"          bun:"created_by"`
}
type ExitByPhoneRequest struct {
	EmployeeID *string `json:"employee_id" form:"employee_id"`
	Latitude   float64 `json:"latitude" form:"latitude"`
	Longitude  float64 `json:"longitude" form:"longitude"`
}

type ExitResponse struct {
	bun.BaseModel `bun:"table:attendance"`

	ID         int       `json:"id" bun:"-"`
	EmployeeID *string   `json:"employee_id" bun:"employee_id"`
	WorkDay    string    `json:"work_day" bun:"work_day"`
	LeaveTime  string    `json:"leave_time,omitempty"`
	TotalHours string    `json:"total_hours" bun:"total_hours"`
	CreatedAt  time.Time `json:"-"          bun:"created_at"`
	CreatedBy  int       `json:"-"          bun:"created_by"`
}
type EnterRequest struct {
	EmployeeID *string `json:"employee_id" form:"employee_id"`
	Latitude   float64 `json:"latitude" form:"latitude"`
	Longitude  float64 `json:"longitude" form:"longitude"`
}

type UpdateRequest struct {
	ID        int    `json:"id" form:"id"`
	WorkDay   string `json:"work_day" form:"work_day"`
	ComeTime  string `json:"come_time" form:"come_time"`
	LeaveTime string `json:"leave_time" form:"leave_time"`
	Status    *bool  `json:"status" form:"status"`
}
type GetStatisticResponse struct {
	TotalEmployee   *int `json:"total_employee" bun:"total_employee"`
	OnTime          *int `json:"ontime" bun:"ontime"`
	Absent          *int `json:"absent" bun:"absent"`
	LateArrival     *int `json:"late_arrival" bun:"late_arrivale"`
	EarlyDepartures *int `json:"early_departures" bun:"early_departures"`
	TimeOff         *int `json:"time_off" bun:"time_off"`
}

type PieChartResponse struct {
	Come   *int `json:"come" bun:"come"`
	Absent *int `json:"absent" bun:"absent"`
}
type GraphRequest struct {
	Month    date.Date
	Interval int
}
type GraphResponse struct {
	Percentage *float64   `json:"percentage" bun:"percentage"`
	WorkDay    *date.Date `json:"work_day" bun:"work_day"`
}
type BarChartResponse struct {
	Department *string  `json:"department" bun:"department"`
	Percentage *float64 `json:"percentage" bun:"percentage"`
}
type Attendance struct {
	ID             int             `json:"id" bun:"id,pk,autoincrement"`
	EmployeeID     *string         `json:"employee_id" bun:"employee_id"`
	ComeTime       string          `json:"come_time,omitempty" bun:"come_time"`
	LeaveTime      string          `json:"leave_time,omitempty" bun:"leave_time"`
	Periods        json.RawMessage `json:"periods"`
	Status         *bool           `json:"status,omitempty" bun:"status"`
	ComeLatitude   *float64        `json:"come_latitude,omitempty" bun:"come_latitude"`
	ComeLongitude  *float64        `json:"come_longitude,omitempty" bun:"come_longitude"`
	LeaveLatitude  *float64        `json:"leave_latitude,omitempty" bun:"leave_latitude"`
	LeaveLongitude *float64        `json:"leave_longitude,omitempty" bun:"leave_longitude"`
	WorkDay        string          `json:"work_day" bun:"work_day"`
	CreatedAt      time.Time       `json:"created_at" bun:"created_at"`
	CreatedBy      int             `json:"created_by" bun:"created_by"`
	UpdatedAt      time.Time       `json:"updated_at" bun:"updated_at"`
	UpdatedBy      int             `json:"updated_by" bun:"updated_by"`
	DeletedAt      *time.Time      `json:"deleted_at,omitempty" bun:"deleted_at"`
}
