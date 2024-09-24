package companyInfo

import (
	"mime/multipart"
	"time"

	"github.com/uptrace/bun"
)

type GetListResponse struct {
	ID           int     `json:"id"`
	Name         *string `json:"name"`
	DepartmentID *int    `json:"department_id"`
	Department   *string `json:"department"`
}

type CreateRequest struct {
	ID          int                   `json:"id" form:"-"`
	CompanyName string                `json:"company_name" form:"company_name"`
	Logo        *multipart.FileHeader `json:"logo" form:"logo"`
	Url         string                `json:"url" form:"url"`
	Latitude    float64               `json:"latitude" form:"latitude"`
	Longitude   float64               `json:"longitude" form:"longitude"`
	CreatedAt   string                `json:"created_at" form:"created_at"`
	UpdatedAt   string                `json:"updated_at" form:"updated_at"`
}

type CreateResponse struct {
	bun.BaseModel `bun:"table:company_info"`

	ID          int       `json:"id" bun:"-"`
	CompanyName string    `json:"company_name" bun:"company_name"`
	Url         string    `json:"url" bun:"url"`
	Latitude    float64   `json:"latitude" bun:"latitude"`
	Longitude   float64   `json:"longitue" bun:"longitude"`
	CreatedAt   time.Time `json:"-" bun:"created_at"`
	CreatedBy   int       `json:"-"          bun:"created_by"`
	UpdatedAt   time.Time `json:"-" bun:"updated_at"`
}

type UpdateRequest struct {
	ID          int                   `json:"id" form:"id"`
	CompanyName string                `json:"company_name" form:"company_name"`
	Logo        *multipart.FileHeader `json:"logo" form:"logo"`
	Url         string                `json:"url" form:"url"`
	Latitude    *float64               `json:"latitude" form:"latitude"`
	Longitude   *float64               `json:"longitude" form:"longitude"`
	StartTime   string                `json:"start_time" form:"start_time"`
	EndTime     string                `json:"end_time" form:"end_time"`
	LateTime    string                `json:"late_time" form:"late_time"`
	OverEndTime string                `json:"over_end_time" form:"over_end_time"`
}
type GetInfoResponse struct {
	bun.BaseModel `bun:"table:company_info"`

	ID          int     `json:"id" `
	CompanyName string  `json:"company_name" `
	Url         string  `json:"url" `
	Latitude    float64 `json:"latitude" `
	Longitude   float64 `json:"longitude" `
	StartTime   string  `json:"start_time" `
	EndTime     string  `json:"end_time" `
	LateTime    string  `json:"late_time" `
	OverEndTime string  `json:"over_end_time" `
}
