package entity

import (
	"time"

	"github.com/uptrace/bun"
)

type Attendance struct {
	bun.BaseModel `bun:"table:attendance"`

	BasicEntity
	DepartmentID *int       `json:"department_id" bun:"department_id"`
	PositionID   *int       `json:"position_id" bun:"position_id"`
	UserID       *int       `json:"user_id" bun:"user_id"`
	ComeTime     *time.Time `json:"come_time" bun:"come_time"`
	LeaveTime    *time.Time `json:"leave_time" bun:"leave_time"`
	Status       *string    `json:"status"   bun:"status"`
}
