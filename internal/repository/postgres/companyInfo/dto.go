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
	Url             string                `json:"url" form:"url"`
	Latitude        *float64              `json:"latitude" form:"latitude"`
	Longitude       *float64              `json:"longitude" form:"longitude"`
	StartTime       string                `json:"start_time" form:"start_time"`
	EndTime         string                `json:"end_time" form:"end_time"`
	LateTime        string                `json:"late_time" form:"late_time"`
	OverEndTime     string                `json:"over_end_time" form:"over_end_time"`
	ComeColor       string                `json:"come_color" form:"come_color"`
	LeaveColor      string                `json:"leave_color" form:"leave_color"`
	ForgetTimeColor string                `json:"forget_time_color" form:"forget_time_color"`
}
type GetInfoResponse struct {
	bun.BaseModel `bun:"table:company_info"`

	ID              int     `json:"id" `
	CompanyName     string  `json:"company_name" `
	Url             string  `json:"url" `
	Latitude        float64 `json:"latitude" `
	Longitude       float64 `json:"longitude" `
	StartTime       string  `json:"start_time" `
	EndTime         string  `json:"end_time" `
	LateTime        string  `json:"late_time" `
	OverEndTime     string  `json:"over_end_time" `
	ComeColor       string  `json:"come_color"`
	LeaveColor      string  `json:"leave_color"`
	ForgetTimeColor string  `json:"forget_time_color"`
}
