package api

import (
	"net/http"

	"github.com/entwico/helm-deployer/domain"
	"github.com/labstack/echo"
)

//ListReleases returns list of releases
func (api *API) ListReleases(ctx echo.Context) error {
	var err error

	items, err := api.services.ReleaseService.ListReleases()
	if err != nil {
		response := &MessageResponse{Message: err.Error()}
		return ctx.JSON(http.StatusInternalServerError, response)
	}
	response := &ListResponse{Page: 1, PageSize: len(items), Total: len(items), Items: items}
	return ctx.JSON(http.StatusOK, response)
}

//UpdateRelease updates release
func (api *API) UpdateRelease(ctx echo.Context) error {
	r := new(domain.ReleaseUpdateRequest)
	if err := ctx.Bind(r); err != nil {
		response := &MessageResponse{Message: err.Error()}
		return ctx.JSON(http.StatusInternalServerError, response)
	}
	if r.Name == "" {
		r.Name = ctx.Param("name")
	}
	err := api.services.ReleaseService.UpdateRelease(r)
	if err != nil {
		response := &MessageResponse{Message: err.Error()}
		return ctx.JSON(http.StatusInternalServerError, response)
	}
	return ctx.JSON(http.StatusOK, "ok")
}
