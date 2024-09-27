package attendance

import (
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
	Date         *string
}

type GetListResponse struct {
	ID           int     `json:"id"`
	DepartmentID *int    `json:"department_id"`
	Department   *string `json:"department"`
	PositionID   *int    `json:"position_id"`
	Position     *string `json:"position"`
	EmployeeID   *string `json:"employee_id"`
	Fullname     *string `json:"full_name"`
	Status       *bool   `json:"status"`
	WorkDay      *string `json:"work_day"`
	ComeTime     *string `json:"come_time,omitempty"`
	LeaveTime    *string `json:"leave_time,omitempty"`
	TotalHours   string  `json:"total_hourse"`
}

type GetDetailByIdResponse struct {
	ID           int     `json:"id"`
	DepartmentID *int    `json:"department_id"`
	Department   *string `json:"department"`
	PositionID   *int    `json:"position_id"`
	Position     *string `json:"position"`
	EmployeeID   *string `json:"employee_id"`
	Fullname     *string `json:"full_name"`
	Status       *bool   `json:"status"`
	WorkDay      *string `json:"work_day"`
	ComeTime     *string `json:"come_time,omitempty"`
	LeaveTime    *string `json:"leave_time,omitempty"`
	TotalHours   string  `json:"total_hours"`
}
type GetHistoryByIdResponse struct {
	EmployeeID *string `json:"employee_id"`
	Fullname   *string `json:"full_name"`
	Status     *bool   `json:"status"`
	WorkDay    *string `json:"work_day"`
	ComeTime   *string `json:"come_time,omitempty"`
	LeaveTime  *string `json:"leave_time,omitempty"`
	TotalHours string  `json:"total_hours"`
}
type GetHistoryByIdRequest struct {
	EmployeeID string     `json:"employee_id"`
	Date       *date.Date `json:"date"`
}

type CreateRequest struct {
	EmployeeID *string `json:"employee_id" form:"employee_id"`
}

type CreateResponse struct {
	bun.BaseModel `bun:"table:attendance"`

	ID         int       `json:"id" bun:"id,autoincrement"` // Ensure auto-increment is understood
	EmployeeID *string   `json:"employee_id" bun:"employee_id"`
	WorkDay    *string   `json:"work_day" bun:"work_day"`
	ComeTime   *string   `json:"come_time" bun:"come_time"`
	LeaveTime  *string   `json:"leave_time" bun:"leave_time"`
	CreatedAt  time.Time `json:"-"          bun:"created_at"`
	CreatedBy  int       `json:"-"          bun:"created_by"`
}
type AttendancePeriod struct {
	bun.BaseModel `bun:"table:attendance_period"`

	ComeTime     *string `json:"come_time" bun:"come_time"`
	AttendanceID *int    `json:"attendance_id" bun:"attendance_id"`
}
type EmployeeResponse struct {
	bun.BaseModel `bun:"table:users"`
	Fullname      *string `json:"full_name" bun:"full_name"`
}
type PeriodsCreate struct {
	bun.BaseModel `bun:"table:attendance_period"`

	ID         int    `json:"id" bun:"-"`
	Attendance int    `json:"attendance_id" bun:"attendance_id"`
	WorkDay    string `json:"work_day" bun:"work_day"`
	ComeTime   string `json:"come_time" bun:"come_time"`
}
type PeriodsUpdate struct {
	bun.BaseModel `bun:"table:attendance_period"`

	ID         int    `json:"id" bun:"-"`
	Attendance int    `json:"attendance_id" bun:"attendance_id"`
	WorkDay    string `json:"work_day" bun:"work_day"`
	LeaveTime  string `json:"leave_time" bun:"leave_time"`
}

type ExitByPhoneRequest struct {
	EmployeeID *string `json:"employee_id" form:"employee_id"`
	Latitude   float64 `json:"latitude" form:"latitude"`
	Longitude  float64 `json:"longitude" form:"longitude"`
}
type EnterRequest struct {
	Latitude   float64 `json:"latitude" form:"latitude"`
	Longitude  float64 `json:"longitude" form:"longitude"`
	EmployeeID *string `json:"employee_id" form:"employee_id"`
}

type UpdateRequest struct {
	ID        int    `json:"id" form:"id"`
	WorkDay   string `json:"work_day" form:"work_day"`
	ComeTime  string `json:"come_time" form:"come_time"`
	LeaveTime string `json:"leave_time" form:"leave_time"`
}
type GetStatisticResponse struct {
	TotalEmployee   *int `json:"total_employee" bun:"total_employee"`
	OnTime          *int `json:"ontime" bun:"ontime"`
	Absent          *int `json:"absent" bun:"absent"`
	LateArrival     *int `json:"late_arrival" bun:"late_arrivale"`
	
	EarlyDepartures *int `json:"early_departures" bun:"early_departures"`
	EarlyCome       *int `json:"early_come" bun:"early_come"`
	OverTime        *int `json:"over_time" bun:"over_time"`
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
	Percentage float64    `json:"percentage" bun:"percentage"`
	WorkDay    *date.Date `json:"work_day" bun:"work_day"`
}
type BarChartResponse struct {
	Department *string  `json:"department" bun:"department"`
	Percentage *float64 `json:"percentage" bun:"percentage"`
}
type Attendance struct {
	ID         int        `json:"id" bun:"id,pk,autoincrement"`
	EmployeeID *string    `json:"employee_id" bun:"employee_id"`
	ComeTime   string     `json:"come_time,omitempty" bun:"come_time"`
	LeaveTime  string     `json:"leave_time,omitempty" bun:"leave_time"`
	Status     *bool      `json:"status,omitempty" bun:"status"`
	WorkDay    string     `json:"work_day" bun:"work_day"`
	CreatedAt  time.Time  `json:"created_at" bun:"created_at"`
	CreatedBy  int        `json:"created_by" bun:"created_by"`
	UpdatedAt  time.Time  `json:"updated_at" bun:"updated_at"`
	UpdatedBy  int        `json:"updated_by" bun:"updated_by"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty" bun:"deleted_at"`
}
