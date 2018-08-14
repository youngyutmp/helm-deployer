package cmd

import (
	"github.com/Sirupsen/logrus"
	"github.com/entwico/helm-deployer/api"
	"github.com/entwico/helm-deployer/conf"
	"github.com/entwico/helm-deployer/domain"
	"github.com/entwico/helm-deployer/service"
	"github.com/pkg/errors"
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
	services, err := configureServices(config)
	if err != nil {
		logrus.Fatal(err)
	}

	apiServer := api.NewAPI(config, services)

	closer.Bind(func() {
		err := apiServer.Stop()
		if err != nil {
			logrus.Fatalf("could not stop API server: %v", err)
		}
	})
	go services.K8SReleaseProvider.Start()
	go func() {
		logrus.Debug("start watching deploy config events")
		services.WebhookDispatcher.StartHandleDeployConfigEvents()
	}()

	logrus.Infof("API started on: %s:%d", config.API.Host, config.API.Port)
	_ = apiServer.Start()
}

func configureServices(config *conf.Config) (*domain.Services, error) {
	db, err := conf.BoltConnect(config)
	if err != nil {
		return nil, errors.Wrap(err, "could not open database")
	}

	webhookRepository, err := service.NewWebhookRepository(db)
	if err != nil {
		return nil, errors.Wrap(err, "could not create webhookRepository")
	}
	chartValuesRepository, err := service.NewChartValuesRepository(db)
	if err != nil {
		return nil, errors.Wrap(err, "could not create chartValuesRepository")
	}

	k8SReleaseProvider, err := service.NewK8SReleaseProvider(config.K8S.ConfigPath)
	if err != nil {
		return nil, errors.Wrap(err, "could not create K8SReleaseProvider")
	}
	services := &domain.Services{
		ChartValuesService:     service.NewChartValuesService(chartValuesRepository),
		ChartRepositoryService: service.NewChartRepositoryService(config.ChartRepository.BaseURL),
		K8SReleaseProvider:     k8SReleaseProvider,
		WebhookService:         service.NewWebhookService(webhookRepository),
	}
	services.HelmService = service.NewHelmService(helm.NewClient(helm.Host(config.Tiller.Host)), services.ChartValuesService, services.ChartRepositoryService)
	nexusProcessor := service.NewNexusProcessor(k8SReleaseProvider)
	gitlabProcessor := service.NewGitlabProcessor(services.WebhookService)
	services.WebhookDispatcher = service.NewWebhookDispatcher(services.HelmService, []domain.WebhookProcessor{gitlabProcessor, nexusProcessor})
	services.ReleaseService = service.NewReleaseService(services.HelmService)

	return services, nil
}
