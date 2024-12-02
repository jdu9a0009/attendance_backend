package entity

import (
	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:users"`

	BasicEntity
	EmployeeID   *string `json:"employee_id"   bun:"employee_id"`
	DepartmentID *int    `json:"department_id"   bun:"department_id"`
	PositionID   *int    `json:"position_id"   bun:"position_id"`
	Password     *string `json:"password"   bun:"password"`
	Role         *string `json:"role"       bun:"role"`
}
