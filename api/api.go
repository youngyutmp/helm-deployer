package api

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/entwico/helm-deployer/conf"
	"github.com/entwico/helm-deployer/domain"
	"github.com/entwico/helm-deployer/embedded"
	"github.com/entwico/helm-deployer/enums"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

// API is the data holder for the API
type API struct {
	config *conf.Config
	log    *logrus.Entry
	echo   *echo.Echo

	services *domain.Services
}

// ListResponse for REST API
type ListResponse struct {
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
	Total    int         `json:"total"`
	Items    interface{} `json:"items"`
}

// MessageResponse for REST API
type MessageResponse struct {
	Status  enums.ResponseStatus `json:"status"`
	Message string               `json:"message"`
	Errors  []ErrorResponseItem  `json:"errors,omitempty"`
}

// ErrorResponseItem for REST API
type ErrorResponseItem struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

// Start will start the API on the specified port
func (api *API) Start() error {
	return api.echo.Start(fmt.Sprintf(":%d", api.config.API.Port))
}

// Stop will shutdown the engine internally
func (api *API) Stop() error {
	logrus.Info("Stopping API server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return api.echo.Shutdown(ctx)
}

// NewAPI will create an api instance that is ready to start
func NewAPI(config *conf.Config, services *domain.Services) *API {
	api := &API{
		config:   config,
		log:      logrus.WithField("component", "api"),
		services: services,
	}

	authConfig := middleware.BasicAuthConfig{Realm: "helm-deployer"}
	authConfig.Validator = func(username, password string, c echo.Context) (bool, error) {
		if username == config.APP.Username && password == config.APP.Password {
			return true, nil
		}
		return false, nil
	}
	skipAuth := config.APP.Username == "" && config.APP.Password == ""
	if skipAuth {
		logrus.Debugf("Basic Auth credentials are not configured")
	}
	authConfig.Skipper = func(c echo.Context) bool {
		if skipAuth {
			return true
		}
		return false
	}
	basicAuth := middleware.BasicAuthWithConfig(authConfig)

	// add the endpoints
	e := echo.New()

	e.HideBanner = true
	e.HTTPErrorHandler = api.handleError

	e.GET("/health", api.Health)
	e.POST("/api/v1/callbacks/gitlab", api.GitlabWebhook)

	g := e.Group("/api/v1", basicAuth)

	// chart repository
	g.GET("/charts", api.ListChartItems)

	// charts
	g.GET("/chart-values", api.ListChartValues)
	g.POST("/chart-values", api.CreateChartValues)
	g.GET("/chart-values/:id", api.GetChartValues)
	g.PUT("/chart-values/:id", api.UpdateChartValues)
	g.DELETE("/chart-values/:id", api.DeleteChartValues)

	// webhooks
	g.GET("/webhooks", api.ListWebhooks)
	g.POST("/webhooks", api.CreateWebhook)
	g.GET("/webhooks/:id", api.GetWebhook)
	g.PUT("/webhooks/:id", api.UpdateWebhook)
	g.DELETE("/webhooks/:id", api.DeleteWebhook)
	g.POST("/webhooks/:id/deploy", api.ForceDeploy)

	// releases
	g.GET("/releases", api.ListReleases)
	g.PUT("/releases/:name", api.UpdateRelease)

	e.GET("/*", api.serveVirtualFS, api.frontend404Fallback)

	api.echo = e
	return api
}

func (api *API) serveVirtualFS(ctx echo.Context) error {
	w, r := ctx.Response(), ctx.Request()
	fileSystem := embedded.FS(false)
	_, err := fileSystem.Open(r.URL.Path)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	fileServer := http.FileServer(fileSystem)
	fileServer.ServeHTTP(w, r)
	return nil
}

func (api *API) frontend404Fallback(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		err := next(c)
		if err != nil {
			e, ok := err.(*echo.HTTPError)
			if ok && e.Code == http.StatusNotFound {
				fileSystem := embedded.FS(false)
				f, _ := fileSystem.Open("/index.html")
				buf := bytes.NewBuffer(nil)
				_, _ = io.Copy(buf, f)
				_ = f.Close()
				return c.HTML(http.StatusOK, string(buf.Bytes()))
			}
		}
		return nil
	}
}

//Health returns application health status
func (api *API) Health(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, map[string]string{"status": "UP"})
}
