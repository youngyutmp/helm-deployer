package cmd

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/entwico/helm-deployer/api"
	"github.com/entwico/helm-deployer/conf"
	"github.com/entwico/helm-deployer/domain"
	"github.com/entwico/helm-deployer/service"
	"github.com/spf13/cobra"
	"github.com/xlab/closer"
	"k8s.io/helm/pkg/helm"
)

var serveCmd = cobra.Command{
	Use:   "serve",
	Short: "Start API server",
	Long:  "Start API server on specified host and port",
	Run: func(cmd *cobra.Command, args []string) {
		execWithConfig(cmd, serve)
	},
}

func serve(config *conf.Config) {
	db, err := conf.BoltConnect(config)
	if err != nil {
		logrus.Fatalf("Error opening database: %v", err)
	}

	helmService := service.NewHelmService(helm.NewClient(helm.Host(config.Tiller.Host)))

	webhookRepository, err := service.NewWebhookRepository(db)
	if err != nil {
		logrus.Fatalf("Can't create webhookRepository: %v", err)
	}
	chartValuesRepository, err := service.NewChartValuesRepository(db)
	if err != nil {
		logrus.Fatalf("Can't create chartValuesRepository: %v", err)
	}
	services := &domain.Services{
		ChartValuesService:     service.NewChartValuesService(chartValuesRepository),
		ChartRepositoryService: service.NewChartRepositoryService(config.ChartRepository.BaseURL),
		ReleaseService:         service.NewReleaseService(helmService),
		WebhookService:         service.NewWebhookService(webhookRepository),
	}

	services.WebhookCallbackService = service.NewGitlabWebhookCallbackService(services.WebhookService, services.ChartRepositoryService, services.ChartValuesService, helmService)

	apiServer := api.NewAPI(config, services)

	l := fmt.Sprintf("%v:%v", config.API.Host, config.API.Port)
	logrus.Infof("API started on: %s", l)

	closer.Bind(func() {
		err := apiServer.Stop()
		if err != nil {
			logrus.Fatalf("can't stop API server: %v", err)
		}
	})

	if err := apiServer.Start(); err != nil {
		logrus.Fatalf("can't start API server: %v", err)
	}
}
