package api

import (
	"net/http"

	"io/ioutil"

	"github.com/Sirupsen/logrus"
	"github.com/entwico/helm-deployer/enums"
	"github.com/labstack/echo"
)

//ProcessWebhook listens to webhooks
func (api *API) ProcessWebhook(c echo.Context) error {
	data, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		response := &MessageResponse{Status: enums.StatusError, Message: err.Error()}
		return c.JSON(http.StatusBadRequest, response)
	}
	defer func() { _ = c.Request().Body.Close() }()
	p, err := api.services.WebhookDispatcher.GetWebhookProcessor(c.Request().Header, data)
	if err != nil {
		response := &MessageResponse{Status: enums.StatusError, Message: err.Error()}
		return c.JSON(http.StatusBadRequest, response)
	}

	go func() {
		if err := p.Process(c.Request().Header, data); err != nil {
			logrus.Warnf("could not process webhook: %v", err)
		}
	}()
	return c.JSON(http.StatusOK, &MessageResponse{Message: "ok"})
}
