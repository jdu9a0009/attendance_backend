package position

import (
	"time"

	"github.com/uptrace/bun"
)

type Filter struct {
	Limit        *int
	Offset       *int
	Page         *int
	Search       *string
	DepartmentID *int
}

type GetListResponse struct {
	ID           int     `json:"id"`
	Name         *string `json:"name"`
	DepartmentID *int    `json:"department_id"`
	Department   *string `json:"department"`
}

type GetDetailByIdResponse struct {
	ID           int     `json:"id"`
	Name         *string `json:"name"`
	DepartmentID *int    `json:"department_id"`
	Department   *string `json:"department"`
}

type CreateRequest struct {
	Name         *string `json:"name" form:"name"`
	DepartmentID *int    `json:"department_id" form:"department_id"`
}

type CreateResponse struct {
	bun.BaseModel `bun:"table:position"`

	ID           int       `json:"id" bun:"-"`
	Name         *string   `json:"name"       bun:"name"`
	DepartmentID *int      `json:"department_id" bun:"department_id"`
	CreatedAt    time.Time `json:"-"          bun:"created_at"`
	CreatedBy    int       `json:"-"          bun:"created_by"`
}

type UpdateRequest struct {
	ID           int     `json:"id" form:"id"`
	Name         *string `json:"name" form:"name"`
	DepartmentID *int    `json:"department_id" bun:"department_id"`
}
