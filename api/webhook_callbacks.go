package api

import (
	"io/ioutil"
	"net/http"

	"github.com/entwico/helm-deployer/conf/logging"
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
	p, err := api.services.WebhookDispatcher.GetWebhookProcessor(c.Request().Context(), c.Request().Header, data)
	if err != nil {
		response := &MessageResponse{Status: enums.StatusError, Message: err.Error()}
		return c.JSON(http.StatusBadRequest, response)
	}

	go func() {
		if err := p.Process(c.Request().Context(), c.Request().Header, data); err != nil {
			logger := logging.FromContext(c.Request().Context())
			logger.WithField("error", err).Warn("could not process webhook")
		}
	}()
	return c.JSON(http.StatusOK, &MessageResponse{Message: "ok"})
}
