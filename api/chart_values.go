package api

import (
	"net/http"

	"github.com/entwico/helm-deployer/enums"
	"github.com/entwico/helm-deployer/service"
	"github.com/labstack/echo"
)

func (api *API) ListChartValues(ctx echo.Context) error {
	var err error

	items, err := api.chartValues.FindAll()
	if err != nil {
		response := &MessageResponse{Status: enums.Error, Message: err.Error()}
		return ctx.JSON(http.StatusInternalServerError, response)
	}
	response := &ListResponse{Page: 1, PageSize: len(items), Total: len(items), Items: items}
	return ctx.JSON(http.StatusOK, response)
}

func (api *API) CreateChartValues(ctx echo.Context) error {
	item := new(service.ChartValues)
	if err := ctx.Bind(item); err != nil {
		response := &MessageResponse{Status: enums.Error, Message: err.Error()}
		return ctx.JSON(http.StatusInternalServerError, response)
	}
	item, err := api.chartValues.Create(item)
	if err != nil {
		response := &MessageResponse{Status: enums.Error, Message: err.Error()}
		return ctx.JSON(http.StatusInternalServerError, response)
	}
	return ctx.JSON(http.StatusCreated, item)

}

func (api *API) GetChartValues(ctx echo.Context) error {
	id := ctx.Param("id")
	item, err := api.chartValues.FindOne(id)

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

func (api *API) UpdateChartValues(ctx echo.Context) error {
	id := ctx.Param("id")
	newItem := new(service.ChartValues)
	if err := ctx.Bind(newItem); err != nil {
		response := &MessageResponse{Status: enums.Error, Message: err.Error()}
		return ctx.JSON(http.StatusBadRequest, response)
	}
	item, err := api.chartValues.Update(id, newItem)
	if err != nil {
		response := &MessageResponse{Status: enums.Error, Message: err.Error()}
		return ctx.JSON(http.StatusInternalServerError, response)
	}
	return ctx.JSON(http.StatusOK, item)

}

func (api *API) DeleteChartValues(ctx echo.Context) error {
	id := ctx.Param("id")
	err := api.chartValues.Delete(id)
	if err != nil {
		response := &MessageResponse{Status: enums.Error, Message: err.Error()}
		return ctx.JSON(http.StatusInternalServerError, response)
	}
	response := &MessageResponse{Message: "item deleted"}
	return ctx.JSON(http.StatusOK, response)
}
