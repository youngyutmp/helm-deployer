package api

import (
	"net/http"

	"github.com/entwico/helm-deployer/service"
	"github.com/labstack/echo"
)

func (api *API) ListReleases(ctx echo.Context) error {
	var err error

	items, err := api.releases.ListReleases()
	if err != nil {
		response := &MessageResponse{Message: err.Error()}
		return ctx.JSON(http.StatusInternalServerError, response)
	}
	response := &ListResponse{Page: 1, PageSize: len(items), Total: len(items), Items: items}
	return ctx.JSON(http.StatusOK, response)
}

func (api *API) UpdateRelease(ctx echo.Context) error {
	r := new(service.ReleaseUpdateRequest)
	if err := ctx.Bind(r); err != nil {
		response := &MessageResponse{Message: err.Error()}
		return ctx.JSON(http.StatusInternalServerError, response)
	}
	if r.Name == "" {
		r.Name = ctx.Param("name")
	}
	err := api.releases.UpdateRelease(r)
	if err != nil {
		response := &MessageResponse{Message: err.Error()}
		return ctx.JSON(http.StatusInternalServerError, response)
	}
	return ctx.JSON(http.StatusOK, "ok")
}
