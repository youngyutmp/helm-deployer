package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/entwico/helm-deployer/enums"
	"github.com/labstack/echo"
)

const headerWebhookGitlab = "X-Gitlab-Event"

func (api *API) GitlabWebhook(ctx echo.Context) error {

	data := new(map[string]interface{})
	if err := ctx.Bind(data); err != nil {
		response := &MessageResponse{Status: enums.Error, Message: err.Error()}
		return ctx.JSON(http.StatusBadRequest, response)
	}
	bytes, e := json.Marshal(data)
	if e != nil {
		response := &MessageResponse{Status: enums.Error, Message: e.Error()}
		return ctx.JSON(http.StatusBadRequest, response)
	}
	webhookType := ctx.Request().Header.Get(headerWebhookGitlab)
	logrus.Info(fmt.Sprintf("Received new gitlab webhook. Type: %s", webhookType))
	logrus.Debug(string(bytes))
	go api.webhookCallbacks.ProcessWebhook(webhookType, bytes)
	response := &MessageResponse{Message: "dispatched"}
	return ctx.JSON(http.StatusOK, response)
}
