package api

import (
	"net/http"

	"github.com/entwico/helm-deployer/domain"
	"github.com/entwico/helm-deployer/enums"
	"github.com/labstack/echo"
)

//ListChartValues returns paginated list of chart values objects
func (api *API) ListChartValues(ctx echo.Context) error {
	var err error

	items, err := api.services.ChartValuesService.FindAll()
	if err != nil {
		response := &MessageResponse{Status: enums.Error, Message: err.Error()}
		return ctx.JSON(http.StatusInternalServerError, response)
	}
	response := &ListResponse{Page: 1, PageSize: len(items), Total: len(items), Items: items}
	return ctx.JSON(http.StatusOK, response)
}

//CreateChartValues creates new ChartValue object
func (api *API) CreateChartValues(ctx echo.Context) error {
	item := new(domain.ChartValues)
	if err := ctx.Bind(item); err != nil {
		response := &MessageResponse{Status: enums.Error, Message: err.Error()}
		return ctx.JSON(http.StatusInternalServerError, response)
	}
	item, err := api.services.ChartValuesService.Create(item)
	if err != nil {
		response := &MessageResponse{Status: enums.Error, Message: err.Error()}
		return ctx.JSON(http.StatusInternalServerError, response)
	}
	return ctx.JSON(http.StatusCreated, item)

}

//GetChartValues returns existing ChartValue object
func (api *API) GetChartValues(ctx echo.Context) error {
	id := ctx.Param("id")
	item, err := api.services.ChartValuesService.FindOne(id)

	if err != nil {
		response := &MessageResponse{Status: enums.Error, Message: err.Error()}
		return ctx.JSON(http.StatusInternalServerError, response)
	}
	if item == nil {
		response := &MessageResponse{Status: enums.Error, Message: "item not found"}
		return ctx.JSON(http.StatusNotFound, response)
	}
	return ctx.JSON(http.StatusOK, item)

}

//UpdateChartValues updates existing ChartValue object
func (api *API) UpdateChartValues(ctx echo.Context) error {
	id := ctx.Param("id")
	newItem := new(domain.ChartValues)
	if err := ctx.Bind(newItem); err != nil {
		response := &MessageResponse{Status: enums.Error, Message: err.Error()}
		return ctx.JSON(http.StatusBadRequest, response)
	}
	item, err := api.services.ChartValuesService.Update(id, newItem)
	if err != nil {
		response := &MessageResponse{Status: enums.Error, Message: err.Error()}
		return ctx.JSON(http.StatusInternalServerError, response)
	}
	return ctx.JSON(http.StatusOK, item)

}

//DeleteChartValues deletes ChartValue object
func (api *API) DeleteChartValues(ctx echo.Context) error {
	id := ctx.Param("id")
	err := api.services.ChartValuesService.Delete(id)
	if err != nil {
		response := &MessageResponse{Status: enums.Error, Message: err.Error()}
		return ctx.JSON(http.StatusInternalServerError, response)
	}
	response := &MessageResponse{Message: "item deleted"}
	return ctx.JSON(http.StatusOK, response)
}
