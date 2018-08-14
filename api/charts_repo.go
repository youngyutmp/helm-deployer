package api

import (
	"net/http"

	"github.com/entwico/helm-deployer/enums"
	"github.com/labstack/echo"
)

//ListChartItems returns list of Charts
func (api *API) ListChartItems(ctx echo.Context) error {
	var err error

	items, err := api.services.ChartRepositoryService.FindAllCharts()
	if err != nil {
		response := &MessageResponse{Status: enums.StatusError, Message: err.Error()}
		return ctx.JSON(http.StatusInternalServerError, response)
	}
	response := &ListResponse{Page: 1, PageSize: len(items), Total: len(items), Items: items}
	return ctx.JSON(http.StatusOK, response)
}
