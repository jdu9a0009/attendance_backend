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
	ID            int     `json:"id"`
	Name          *string `json:"name"`
	DisplayNumber int     `json:"display_number"`
}

type GetDetailByIdResponse struct {
	ID            int     `json:"id"`
	Name          *string `json:"name" form:"name"`
	DisplayNumber int     `json:"display_number"`
}

type CreateRequest struct {
	Name          *string `json:"name" form:"name"`
	DisplayNumber int     `json:"display_number" form:"display_number"`
}

type CreateResponse struct {
	bun.BaseModel `bun:"table:department"`

	ID int `json:"id" bun:"-"`

	Name          *string `json:"name"       bun:"name"`
	DisplayNumber int     `json:"display_number" bun:"display_number"`

	CreatedAt time.Time `json:"-"          bun:"created_at"`
	CreatedBy int       `json:"-"          bun:"created_by"`
}

type UpdateRequest struct {
	ID   int     `json:"id" form:"id"`
	Name *string `json:"name" form:"name"`
	DisplayNumber int     `json:"display_number" form:"display_number"`

}
