package api

import (
	"fmt"
	"net/http"

	"github.com/entwico/helm-deployer/domain"
	"github.com/entwico/helm-deployer/enums"
	"github.com/labstack/echo"
)

//ListWebhooks returns a list of Webhook objects
func (api *API) ListWebhooks(ctx echo.Context) error {
	var err error

	items, err := api.services.WebhookService.FindAll()
	if err != nil {
		response := &MessageResponse{Status: enums.Error, Message: err.Error()}
		return ctx.JSON(http.StatusInternalServerError, response)
	}
	response := &ListResponse{Page: 1, PageSize: len(items), Total: len(items), Items: items}
	return ctx.JSON(http.StatusOK, response)
}

//CreateWebhook creates new WebHook
func (api *API) CreateWebhook(ctx echo.Context) error {
	item := new(domain.Webhook)
	if err := ctx.Bind(item); err != nil {
		response := &MessageResponse{Status: enums.Error, Message: err.Error()}
		return ctx.JSON(http.StatusInternalServerError, response)
	}
	item, err := api.services.WebhookService.Create(item)
	if err != nil {
		response := &MessageResponse{Status: enums.Error, Message: err.Error()}
		return ctx.JSON(http.StatusInternalServerError, response)
	}
	return ctx.JSON(http.StatusCreated, item)

}

//GetWebhook returns existing Webhook
func (api *API) GetWebhook(ctx echo.Context) error {
	id := ctx.Param("id")
	item, err := api.services.WebhookService.FindOne(id)

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

//UpdateWebhook updates Webhook
func (api *API) UpdateWebhook(ctx echo.Context) error {
	id := ctx.Param("id")
	newItem := new(domain.Webhook)
	if err := ctx.Bind(newItem); err != nil {
		response := &MessageResponse{Status: enums.Error, Message: err.Error()}
		return ctx.JSON(http.StatusBadRequest, response)
	}
	item, err := api.services.WebhookService.Update(id, newItem)
	if err != nil {
		response := &MessageResponse{Status: enums.Error, Message: err.Error()}
		return ctx.JSON(http.StatusInternalServerError, response)
	}
	return ctx.JSON(http.StatusOK, item)

}

//DeleteWebhook deletes Webhook
func (api *API) DeleteWebhook(ctx echo.Context) error {
	id := ctx.Param("id")
	err := api.services.WebhookService.Delete(id)
	if err != nil {
		response := &MessageResponse{Status: enums.Error, Message: err.Error()}
		return ctx.JSON(http.StatusInternalServerError, response)
	}
	response := &MessageResponse{Message: "item deleted"}
	return ctx.JSON(http.StatusOK, response)
}

//ForceDeploy forces chart redeploy
func (api *API) ForceDeploy(ctx echo.Context) error {
	id := ctx.Param("id")
	w, err := api.services.WebhookService.FindOne(id)
	if err != nil {
		response := &MessageResponse{Status: enums.Error, Message: err.Error()}
		return ctx.JSON(http.StatusInternalServerError, response)
	}
	if w == nil {
		response := &MessageResponse{Status: enums.Error, Message: fmt.Sprintf("webhook %s not found", id)}
		return ctx.JSON(http.StatusNotFound, response)
	}
	err = api.services.WebhookCallbackService.DeployChart(w.DeployConfig)
	if err != nil {
		response := &MessageResponse{Status: enums.Error, Message: err.Error()}
		return ctx.JSON(http.StatusInternalServerError, response)
	}
	response := &MessageResponse{Message: fmt.Sprintf("deploy for %s dispatched", w.DeployConfig.ReleaseName)}
	return ctx.JSON(http.StatusOK, response)
}
