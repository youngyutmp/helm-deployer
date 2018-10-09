package api

import (
	"net/http"

	"github.com/entwico/helm-deployer/domain"
	"github.com/labstack/echo"
)

//ListReleases returns list of releases
func (api *API) ListReleases(c echo.Context) error {
	var err error

	items, err := api.services.ReleaseService.ListReleases(c.Request().Context())
	if err != nil {
		response := &MessageResponse{Message: err.Error()}
		return c.JSON(http.StatusInternalServerError, response)
	}
	response := &ListResponse{Page: 1, PageSize: len(items), Total: len(items), Items: items}
	return c.JSON(http.StatusOK, response)
}

//UpdateRelease updates release
func (api *API) UpdateRelease(c echo.Context) error {
	r := new(domain.ReleaseUpdateRequest)
	if err := c.Bind(r); err != nil {
		response := &MessageResponse{Message: err.Error()}
		return c.JSON(http.StatusInternalServerError, response)
	}
	if r.Name == "" {
		r.Name = c.Param("name")
	}
	err := api.services.ReleaseService.UpdateRelease(c.Request().Context(), r)
	if err != nil {
		response := &MessageResponse{Message: err.Error()}
		return c.JSON(http.StatusInternalServerError, response)
	}
	return c.JSON(http.StatusOK, "ok")
}
