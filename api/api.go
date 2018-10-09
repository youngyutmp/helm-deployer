package api

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/entwico/helm-deployer/conf"
	"github.com/entwico/helm-deployer/conf/logging"
	"github.com/entwico/helm-deployer/domain"
	"github.com/entwico/helm-deployer/embedded"
	"github.com/entwico/helm-deployer/enums"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
)

// API is the data holder for the API
type API struct {
	config *conf.Config
	log    *log.Entry
	echo   *echo.Echo

	// Services used by the API
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
	api.log.Info("stopping API server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return api.echo.Shutdown(ctx)
}

// NewAPI will create an api instance that is ready to start
func NewAPI(config *conf.Config, services *domain.Services) *API {
	logger := config.LogConfig.Logger
	api := &API{
		config:   config,
		log:      logger.WithField("component", "api"),
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
		api.log.Info("basic auth credentials are not configured")
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
	e.HidePort = true

	e.HTTPErrorHandler = api.handleError
	api.echo = e

	e.GET("/health", api.Health)
	e.POST("/api/v1/callbacks", api.ProcessWebhook)
	e.POST("/api/v1/callbacks/:name", api.ProcessWebhook)

	g := e.Group("/api/v1", basicAuth)
	g.Use(api.setupRequest)

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

	return api
}

func (api *API) setupRequest(f echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		req := ctx.Request()
		logger := api.log.WithFields(log.Fields{
			"method":     req.Method,
			"path":       req.URL.Path,
			"request_id": uuid.NewRandom(),
		})

		rqCtx := logging.NewContextWithLogger(req.Context(), logger)
		ctx.SetRequest(ctx.Request().WithContext(rqCtx))

		startTime := time.Now()
		defer func() {
			rsp := ctx.Response()
			logger.WithFields(log.Fields{
				"status_code":   rsp.Status,
				"runtime_milli": time.Since(startTime).Nanoseconds() / (int64(time.Millisecond) / int64(time.Nanosecond)),
			}).Debug("request finished")
		}()

		logger.WithFields(log.Fields{
			"user_agent": req.UserAgent(),
		}).Debug("request started")

		// we have to do this b/c if not the final error handler will not
		// in the chain of middleware. It will be called after meaning that the
		// response won't be set properly.
		err := f(ctx)
		if err != nil {
			ctx.Error(err)
		}
		return err
	}
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
