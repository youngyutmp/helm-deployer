package api

import (
	"fmt"
	"net/http"

	"github.com/entwico/helm-deployer/domain"
	"github.com/entwico/helm-deployer/enums"
	"github.com/labstack/echo"
)

//ListWebhooks returns a list of Webhook objects
func (api *API) ListWebhooks(c echo.Context) error {
	var err error

	items, err := api.services.WebhookService.FindAll()
	if err != nil {
		response := &MessageResponse{Status: enums.StatusError, Message: err.Error()}
		return c.JSON(http.StatusInternalServerError, response)
	}
	response := &ListResponse{Page: 1, PageSize: len(items), Total: len(items), Items: items}
	return c.JSON(http.StatusOK, response)
}

//CreateWebhook creates new WebHook
func (api *API) CreateWebhook(c echo.Context) error {
	item := new(domain.Webhook)
	if err := c.Bind(item); err != nil {
		response := &MessageResponse{Status: enums.StatusError, Message: err.Error()}
		return c.JSON(http.StatusInternalServerError, response)
	}
	item, err := api.services.WebhookService.Create(item)
	if err != nil {
		response := &MessageResponse{Status: enums.StatusError, Message: err.Error()}
		return c.JSON(http.StatusInternalServerError, response)
	}
	return c.JSON(http.StatusCreated, item)

}

//GetWebhook returns existing Webhook
func (api *API) GetWebhook(c echo.Context) error {
	id := c.Param("id")
	item, err := api.services.WebhookService.FindOne(id)

	if err != nil {
		response := &MessageResponse{Status: enums.StatusError, Message: err.Error()}
		return c.JSON(http.StatusInternalServerError, response)
	}
	if item == nil {
		response := &MessageResponse{Status: enums.StatusError, Message: "item not found"}
		return c.JSON(http.StatusNotFound, response)
	}
	return c.JSON(http.StatusOK, item)

}

//UpdateWebhook updates Webhook
func (api *API) UpdateWebhook(c echo.Context) error {
	id := c.Param("id")
	newItem := new(domain.Webhook)
	if err := c.Bind(newItem); err != nil {
		response := &MessageResponse{Status: enums.StatusError, Message: err.Error()}
		return c.JSON(http.StatusBadRequest, response)
	}
	item, err := api.services.WebhookService.Update(id, newItem)
	if err != nil {
		response := &MessageResponse{Status: enums.StatusError, Message: err.Error()}
		return c.JSON(http.StatusInternalServerError, response)
	}
	return c.JSON(http.StatusOK, item)

}

//DeleteWebhook deletes Webhook
func (api *API) DeleteWebhook(c echo.Context) error {
	id := c.Param("id")
	err := api.services.WebhookService.Delete(id)
	if err != nil {
		response := &MessageResponse{Status: enums.StatusError, Message: err.Error()}
		return c.JSON(http.StatusInternalServerError, response)
	}
	response := &MessageResponse{Message: "item deleted"}
	return c.JSON(http.StatusOK, response)
}

//ForceDeploy forces chart redeploy
func (api *API) ForceDeploy(c echo.Context) error {
	id := c.Param("id")
	w, err := api.services.WebhookService.FindOne(id)
	if err != nil {
		response := &MessageResponse{Status: enums.StatusError, Message: err.Error()}
		return c.JSON(http.StatusInternalServerError, response)
	}
	if w == nil {
		response := &MessageResponse{Status: enums.StatusError, Message: fmt.Sprintf("webhook %s not found", id)}
		return c.JSON(http.StatusNotFound, response)
	}

	err = api.services.HelmService.DeployChart(c.Request().Context(), w.DeployConfig)
	if err != nil {
		response := &MessageResponse{Status: enums.StatusError, Message: err.Error()}
		return c.JSON(http.StatusInternalServerError, response)
	}
	response := &MessageResponse{Message: fmt.Sprintf("deploy for %s dispatched", w.DeployConfig.ReleaseName)}
	return c.JSON(http.StatusOK, response)
}
