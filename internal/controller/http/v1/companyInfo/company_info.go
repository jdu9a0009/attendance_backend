package companyInfo

import (
	"attendance/backend/foundation/web"
	"attendance/backend/internal/repository/postgres/companyInfo"
	"attendance/backend/internal/service"
	"net/http"
	"reflect"
)

type Controller struct {
	companyInfo CompanyInfo
}

const companyDir = "company_info" 

func NewController(companyInfo CompanyInfo) *Controller {
	return &Controller{companyInfo}
}

func (uc Controller) UpdateAll(c *web.Context) error {
	id := c.GetParam(reflect.Int, "id").(int)

	if err := c.ValidParam(); err != nil {
		return c.RespondError(err)
	}

	var request companyInfo.UpdateRequest

	if err := c.BindFunc(&request, "company_name", "latitude", "longitude"); err != nil {
		return c.RespondError(err)
	}

	request.ID = id

	// Check if image exists in the request
	if request.Logo != nil {
		path, err := service.Upload(request.Logo, companyDir)
		if err != nil {
			return c.RespondError(err)
		}
		request.Url = path
	}

	err := uc.companyInfo.UpdateAll(c.Ctx, request)
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

	response, err := uc.companyInfo.GetInfo(c.Ctx)
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
