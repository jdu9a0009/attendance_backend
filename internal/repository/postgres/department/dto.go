package department

import (
	"time"

	"github.com/uptrace/bun"
)

type Filter struct {
	Limit  *int
	Offset *int
	Page   *int
	Search *string
}

type GetListResponse struct {
	ID   int     `json:"id"`
	Name *string `json:"name"`
}

type GetDetailByIdResponse struct {
	ID   int     `json:"id"`
	Name *string `json:"name" form:"name"`
}

type CreateRequest struct {
	Name *string `json:"name" form:"name"`
}

type CreateResponse struct {
	bun.BaseModel `bun:"table:department"`

	ID int `json:"id" bun:"-"`

	Name      *string   `json:"name"       bun:"name"`
	CreatedAt time.Time `json:"-"          bun:"created_at"`
	CreatedBy int       `json:"-"          bun:"created_by"`
}

type UpdateRequest struct {
	ID   int     `json:"id" form:"id"`
	Name *string `json:"name" form:"name"`
}
