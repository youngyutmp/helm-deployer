package api

import (
	"net/http"

	"github.com/entwico/helm-deployer/enums"
	"github.com/labstack/echo"
)

//ListChartItems returns list of Charts
func (api *API) ListChartItems(c echo.Context) error {
	var err error

	items, err := api.services.ChartRepositoryService.FindAllCharts(c.Request().Context())
	if err != nil {
		response := &MessageResponse{Status: enums.StatusError, Message: err.Error()}
		return c.JSON(http.StatusInternalServerError, response)
	}
	response := &ListResponse{Page: 1, PageSize: len(items), Total: len(items), Items: items}
	return c.JSON(http.StatusOK, response)
}
