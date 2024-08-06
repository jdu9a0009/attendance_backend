package user

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

type SignInRequest struct {
	EmployeeID string `json:"employee_id" form:"employee_id"`
	Password   string `json:"password" form:"password"`
}

type AuthClaims struct {
	ID   int
	Role string
	Type string
}

type RefreshTokenRequest struct {
	AccessToken  string `json:"access_token" form:"access_token"`
	RefreshToken string `json:"refresh_token" form:"refresh_token"`
}

type GetListResponse struct {
	ID           int     `json:"id"`
	EmployeeID   *string `json:"employee_id"`
	FullName     *string `json:"full_name"`
	DepartmentID *int    `json:"department_id"`
	Department   *string `json:"department"`
	PositionID   *int    `json:"position_id"`
	Position     *string `json:"position"`
	Phone        *string `json:"phone"`
	Email        *string `json:"email"`
}

type GetDetailByIdResponse struct {
	ID           int     `json:"id"`
	EmployeeID   *string `json:"employee_id"`
	FullName     *string `json:"full_name"`
	DepartmentID *int    `json:"department_id"`
	Department   *string `json:"department"`
	PositionID   *int    `json:"position_id"`
	Position     *string `json:"position"`
	Phone        *string `json:"phone"`
	Email        *string `json:"email"`
}

type CreateRequest struct {
	EmployeeID   *string `json:"employee_id"   form:"employee_id"`
	Password     *string `json:"password"   form:"password"`
	Role         *string `json:"role" form:"role"`
	FullName     *string `json:"full_name"  form:"full_name"`
	DepartmentID *int    `json:"department_id" form:"department_id"`
	PositionID   *int    `json:"position_id" form:"position_id"`
	Phone        *string `json:"phone" form:"phone"`
	Email        *string `json:"email" form:"email"`
}

type CreateResponse struct {
	bun.BaseModel `bun:"table:users"`

	ID           int       `json:"id" bun:"-"`
	EmployeeID   *string   `json:"employee_id"   bun:"employee_id"`
	Password     *string   `json:"-"   bun:"password"`
	Role         *string   `json:"role" bun:"role"`
	FullName     *string   `json:"full_name"  bun:"full_name"`
	DepartmentID *int      `json:"department_id" bun:"department_id"`
	PositionID   *int      `json:"position_id" bun:"position_id"`
	Phone        *string   `json:"phone" bun:"phone"`
	Email        *string   `json:"email" bun:"email"`
	CreatedAt    time.Time `json:"-"          bun:"created_at"`
	CreatedBy    int       `json:"-"          bun:"created_by"`
}

type UpdateRequest struct {
	ID           int     `json:"id" form:"id"`
	EmployeeID   *string `json:"employee_id"   form:"employee_id"`
	Password     *string `json:"password"   form:"password"`
	Role         *string `json:"role"       form:"role"`
	FullName     *string `json:"full_name"  form:"full_name"`
	DepartmentID *int    `json:"department_id" form:"department_id"`
	PositionID   *int    `json:"position_id" form:"position_id"`
	Phone        *string `json:"phone" form:"phone"`
	Email        *string `json:"email" form:"email"`
}
