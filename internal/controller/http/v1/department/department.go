package department

import (
	"net/http"
	"reflect"
	"university-backend/foundation/web"
	"university-backend/internal/repository/postgres/department"
)

type Controller struct {
	department Department
}

func NewController(department Department) *Controller {
	return &Controller{department}
}

// department

func (uc Controller) GetList(c *web.Context) error {
	var filter department.Filter

	if limit, ok := c.GetQueryFunc(reflect.Int, "limit").(*int); ok {
		filter.Limit = limit
	}
	if offset, ok := c.GetQueryFunc(reflect.Int, "offset").(*int); ok {
		filter.Offset = offset
	}
	if page, ok := c.GetQueryFunc(reflect.Int, "page").(*int); ok {
		filter.Page = page
	}
	if search, ok := c.GetQueryFunc(reflect.String, "search").(*string); ok {
		filter.Search = search
	}

	if err := c.ValidQuery(); err != nil {
		return c.RespondError(err)
	}

	list, count, err := uc.department.GetList(c.Ctx, filter)
	if err != nil {
		return c.RespondError(err)
	}

	return c.Respond(map[string]interface{}{
		"data": map[string]interface{}{
			"results": list,
			"count":   count,
		},
		"status": true,
	}, http.StatusOK)
}

func (uc Controller) GetDetailById(c *web.Context) error {
	id := c.GetParam(reflect.Int, "id").(int)

	if err := c.ValidParam(); err != nil {
		return c.RespondError(err)
	}

	response, err := uc.department.GetDetailById(c.Ctx, id)
	if err != nil {
		return c.RespondError(err)
	}

	return c.Respond(map[string]interface{}{
		"data":   response,
		"status": true,
	}, http.StatusOK)
}

func (uc Controller) Create(c *web.Context) error {
	var request department.CreateRequest

	if err := c.BindFunc(&request, "Name"); err != nil {
		return c.RespondError(err)
	}

	response, err := uc.department.Create(c.Ctx, request)
	if err != nil {
		return c.RespondError(err)
	}

	return c.Respond(map[string]interface{}{
		"data":   response,
		"status": true,
	}, http.StatusOK)
}

func (uc Controller) UpdateAll(c *web.Context) error {
	id := c.GetParam(reflect.Int, "id").(int)

	if err := c.ValidParam(); err != nil {
		return c.RespondError(err)
	}

	var request department.UpdateRequest

	if err := c.BindFunc(&request, "Name"); err != nil {
		return c.RespondError(err)
	}

	request.ID = id

	err := uc.department.UpdateAll(c.Ctx, request)
	if err != nil {
		return c.RespondError(err)
	}

	return c.Respond(map[string]interface{}{
		"data":   "ok!",
		"status": true,
	}, http.StatusOK)
}

func (uc Controller) UpdateColumns(c *web.Context) error {
	id := c.GetParam(reflect.Int, "id").(int)

	if err := c.ValidParam(); err != nil {
		return c.RespondError(err)
	}

	var request department.UpdateRequest

	if err := c.BindFunc(&request); err != nil {
		return c.RespondError(err)
	}

	request.ID = id

	err := uc.department.UpdateColumns(c.Ctx, request)
	if err != nil {
		return c.RespondError(err)
	}

	return c.Respond(map[string]interface{}{
		"data":   "ok!",
		"status": true,
	}, http.StatusOK)
}

func (uc Controller) Delete(c *web.Context) error {
	id := c.GetParam(reflect.Int, "id").(int)

	if err := c.ValidParam(); err != nil {
		return c.RespondError(err)
	}

	err := uc.department.Delete(c.Ctx, id)
	if err != nil {
		return c.RespondError(err)
	}

	return c.Respond(map[string]interface{}{
		"data":   "ok!",
		"status": true,
	}, http.StatusOK)
}
