package entity

import (
	"github.com/uptrace/bun"
)

type Position struct {
	bun.BaseModel `bun:"table:position"`

	BasicEntity
	Name         *string `json:"name"     bun:"name"`
	DepartmentID *int    `json:"department_id" bun:"department_id"`
}
