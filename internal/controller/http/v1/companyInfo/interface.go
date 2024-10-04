package companyInfo

import (
	"context"
	"attendance/backend/internal/repository/postgres/companyInfo"
)

type CompanyInfo interface {
	UpdateAll(ctx context.Context, request companyInfo.UpdateRequest) error
	GetInfo(ctx context.Context) (companyInfo.GetInfoResponse, error)
}
