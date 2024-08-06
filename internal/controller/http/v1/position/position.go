package position

import (
	"net/http"
	"reflect"
	"university-backend/foundation/web"
	"university-backend/internal/repository/postgres/position"
)

type Controller struct {
	position Position
}

func NewController(position Position) *Controller {
	return &Controller{position}
}

// position

func (uc Controller) GetList(c *web.Context) error {
	var filter position.Filter

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
	if departmentId, ok := c.GetQueryFunc(reflect.Int, "department_id").(*int); ok {
		filter.DepartmentID = departmentId
	}

	if err := c.ValidQuery(); err != nil {
		return c.RespondError(err)
	}

	list, count, err := uc.position.GetList(c.Ctx, filter)
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

	response, err := uc.position.GetDetailById(c.Ctx, id)
	if err != nil {
		return c.RespondError(err)
	}

	return c.Respond(map[string]interface{}{
		"data":   response,
		"status": true,
	}, http.StatusOK)
}

func (uc Controller) Create(c *web.Context) error {
	var request position.CreateRequest
	if err := c.BindFunc(&request, "Name"); err != nil {
		return c.RespondError(err)
	}

	response, err := uc.position.Create(c.Ctx, request)
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

	var request position.UpdateRequest

	if err := c.BindFunc(&request, "Name"); err != nil {
		return c.RespondError(err)
	}

	request.ID = id

	err := uc.position.UpdateAll(c.Ctx, request)
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

	var request position.UpdateRequest

	if err := c.BindFunc(&request); err != nil {
		return c.RespondError(err)
	}

	request.ID = id

	err := uc.position.UpdateColumns(c.Ctx, request)
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

	err := uc.position.Delete(c.Ctx, id)
	if err != nil {
		return c.RespondError(err)
	}

	return c.Respond(map[string]interface{}{
		"data":   "ok!",
		"status": true,
	}, http.StatusOK)
}
