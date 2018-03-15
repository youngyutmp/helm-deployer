package api

import (
	"net/http"

	"github.com/entwico/helm-deployer/enums"
	"github.com/entwico/helm-deployer/service"
	"github.com/labstack/echo"
)

func (api *API) ListWebhooks(ctx echo.Context) error {
	var err error

	items, err := api.webhooks.FindAll()
	if err != nil {
		response := &MessageResponse{Status: enums.Error, Message: err.Error()}
		return ctx.JSON(http.StatusInternalServerError, response)
	}
	response := &ListResponse{Page: 1, PageSize: len(items), Total: len(items), Items: items}
	return ctx.JSON(http.StatusOK, response)
}

func (api *API) CreateWebhook(ctx echo.Context) error {
	item := new(service.Webhook)
	if err := ctx.Bind(item); err != nil {
		response := &MessageResponse{Status: enums.Error, Message: err.Error()}
		return ctx.JSON(http.StatusInternalServerError, response)
	}
	item, err := api.webhooks.Create(item)
	if err != nil {
		response := &MessageResponse{Status: enums.Error, Message: err.Error()}
		return ctx.JSON(http.StatusInternalServerError, response)
	}
	return ctx.JSON(http.StatusCreated, item)

}

func (api *API) GetWebhook(ctx echo.Context) error {
	id := ctx.Param("id")
	item, err := api.webhooks.FindOne(id)

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

func (api *API) UpdateWebhook(ctx echo.Context) error {
	id := ctx.Param("id")
	newItem := new(service.Webhook)
	if err := ctx.Bind(newItem); err != nil {
		response := &MessageResponse{Status: enums.Error, Message: err.Error()}
		return ctx.JSON(http.StatusBadRequest, response)
	}
	item, err := api.webhooks.Update(id, newItem)
	if err != nil {
		response := &MessageResponse{Status: enums.Error, Message: err.Error()}
		return ctx.JSON(http.StatusInternalServerError, response)
	}
	return ctx.JSON(http.StatusOK, item)

}

func (api *API) DeleteWebhook(ctx echo.Context) error {
	id := ctx.Param("id")
	err := api.webhooks.Delete(id)
	if err != nil {
		response := &MessageResponse{Status: enums.Error, Message: err.Error()}
		return ctx.JSON(http.StatusInternalServerError, response)
	}
	response := &MessageResponse{Message: "item deleted"}
	return ctx.JSON(http.StatusOK, response)
}
