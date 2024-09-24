package companyInfo

import (
	"net/http"
	"reflect"
	"university-backend/foundation/web"
	"university-backend/internal/repository/postgres/companyInfo"
	"university-backend/internal/service"
)

type Controller struct {
	companyInfo CompanyInfo
}

const companyDir = "company_info"

func NewController(companyInfo CompanyInfo) *Controller {
	return &Controller{companyInfo}
}

func (uc Controller) Create(c *web.Context) error {
	var request companyInfo.CreateRequest
	if err := c.BindFunc(&request); err != nil {
		return c.RespondError(err)
	}

	// Log path initialization
	var logPath string
	logPath = logPath + "->Create"

	// Check if image exists in the request
	if request.Logo != nil {
		path, err := service.Upload(request.Logo, companyDir)
		if err != nil {
			return c.RespondError(err)
		}
		request.Url = path
	}

	// Create company info
	response, err := uc.companyInfo.Create(c.Ctx, request)
	if err != nil {
		return c.RespondError(err)
	}

	// Respond with the created data and status
	return c.Respond(map[string]interface{}{
		"data":   response,
		"status": true,
	}, http.StatusOK)
}

func (uc Controller) UpdateColumns(c *web.Context) error {
	id := c.GetParam(reflect.Int, "id").(int)

	if err := c.ValidParam(); err != nil {
		return c.RespondError(err)
	}

	var request companyInfo.UpdateRequest

	if err := c.BindFunc(&request); err != nil {
		return c.RespondError(err)
	}

	request.ID = id

	err := uc.companyInfo.UpdateColumns(c.Ctx, request)
	if err != nil {
		return c.RespondError(err)
	}

	return c.Respond(map[string]interface{}{
		"data":   "ok!",
		"status": true,
	}, http.StatusOK)
}
func (uc Controller) GetInfo(c *web.Context) error {


	if err := c.ValidQuery(); err != nil {
		return c.RespondError(err)
	}

	response,  err := uc.companyInfo.GetInfo(c.Ctx, )
	if err != nil {
		return c.RespondError(err)
	}

	return c.Respond(map[string]interface{}{
		"data": map[string]interface{}{
			"results": response,
		},
		"status": true,
	}, http.StatusOK)
}
