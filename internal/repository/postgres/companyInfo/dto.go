package companyInfo

import (
	"mime/multipart"

	"github.com/uptrace/bun"
)

type GetListResponse struct {
	ID           int     `json:"id"`
	Name         *string `json:"name"`
	DepartmentID *int    `json:"department_id"`
	Department   *string `json:"department"`
}

type UpdateRequest struct {
	ID              int                   `json:"id" form:"id"`
	CompanyName     string                `json:"company_name" form:"company_name"`
	Logo            *multipart.FileHeader `json:"logo" form:"logo"`
	Bold        bool                  `json:"bold" form:"bold"`
	Url             string                `json:"url" form:"url"`
	Latitude        *float64              `json:"latitude" form:"latitude"`
	Longitude       *float64              `json:"longitude" form:"longitude"`
	Radius          float64               `json:"radius" form:"radius"`
	StartTime       string                `json:"start_time" form:"start_time"`
	EndTime         string                `json:"end_time" form:"end_time"`
	LateTime        string                `json:"late_time" form:"late_time"`
	OverEndTime     string                `json:"over_end_time" form:"over_end_time"`
	ComeTimeColor   string                `json:"come_time_color" form:"come_time_color"`
	LeaveTimeColor  string                `json:"leave_time_color" form:"leave_time_color"`
	ForgetTimeColor string                `json:"forget_time_color" form:"forget_time_color"`
	PresentColor    string                `json:"present_color" form:"present_color"`
	AbsentColor     string                `json:"absent_color" form:"absent_color"`
	NewPresentColor string                `json:"new_present_color" form:"new_present_color"`
	NewAbsentColor  string                `json:"new_absent_color" form:"new_absent_color"`
}
type GetInfoResponse struct {
	bun.BaseModel `bun:"table:company_info"`

	ID              int     `json:"id"`
	CompanyName     string  `json:"company_name"`
	Url             string  `json:"url"`
	Bold        bool    `json:"bold"`
	Latitude        float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`
	Radius          float64 `json:"radius"`
	StartTime       string  `json:"start_time"`
	EndTime         string  `json:"end_time"`
	LateTime        string  `json:"late_time"`
	OverEndTime     string  `json:"over_end_time"`
	ComeColor       string  `json:"come_time_color" bun:"come_time_color"`
	LeaveColor      string  `json:"leave_time_color" bun:"leave_time_color"`
	ForgetTimeColor string  `json:"forget_time_color" bun:"forget_time_color"`
	PresentColor    string  `json:"present_color" bun:"present_color"`
	AbsentColor     string  `json:"absent_color" bun:"absent_color"`
	NewPresentColor string  `json:"new_present_color" bun:"new_present_color"`
	NewAbsentColor  string  `json:"new_absent_color" bun:"new_absent_color"`
}

type GetAttendanceColorResponse struct {
	bun.BaseModel `bun:"table:company_info"`

	ComeColor       string `json:"come_time_color" bun:"come_time_color"`
	LeaveColor      string `json:"leave_time_color" bun:"leave_time_color"`
	ForgetTimeColor string `json:"forget_time_color" bun:"forget_time_color"`
	PresentColor    string `json:"present_color" bun:"present_color"`
	AbsentColor     string `json:"absent_color" bun:"absent_color"`
}
type GetNewTableColorResponse struct {
	bun.BaseModel   `bun:"table:company_info"`
	NewPresentColor string `json:"new_present_color" bun:"new_present_color"`
	NewAbsentColor  string `json:"new_absent_color" bun:"new_absent_color"`
}
