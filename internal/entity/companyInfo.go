package entity

import (
	"mime/multipart"

	"github.com/uptrace/bun"
)

type CompanyInfo struct {
	bun.BaseModel `bun:"table:company_info"`

	BasicEntity
	ID          int                   `json:"id" bun:"-"`
	CompanyName string                `json:"company_name" bun:"company_name"`
	Image       *multipart.FileHeader `json:"image" bun:"image"`
	Url         string                `json:"url" bun:"-"`
	StartTime   string                `json:"start_time" bun:"start_time"`
	EndTime     string                `json:"end_time" bun:"end_time"`
	LateTime    string                `json:"late_time" bun:"late_time"`
	OverEndTime string                `json:"over_end_time" bun:"over_end_time"`
	Latitude    float64               `json:"latitude" bun:"latitude"`
	Longitude   float64               `json:"longitude" bun:"longitude"`
}
