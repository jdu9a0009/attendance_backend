package companyInfo

import (
	"context"
	"university-backend/internal/repository/postgres/companyInfo"
)

type CompanyInfo interface {
	Create(ctx context.Context, request companyInfo.CreateRequest) (companyInfo.CreateResponse, error)
	UpdateColumns(ctx context.Context, request companyInfo.UpdateRequest) error
	GetInfo(ctx context.Context) (companyInfo.GetInfoResponse, error)
}
