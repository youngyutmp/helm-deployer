package service

import (
	"context"
	"net/http"
	"sync"

	"github.com/entwico/helm-deployer/conf/logging"
	"github.com/entwico/helm-deployer/domain"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type webhookProcessor struct {
	helmService domain.HelmService
	processors  []domain.WebhookProcessor
}

//NewWebhookDispatcher returns a new instance of WebhookProcessor
func NewWebhookDispatcher(helmService domain.HelmService, processors []domain.WebhookProcessor) domain.WebhookDispatcher {
	return &webhookProcessor{helmService: helmService, processors: processors}
}

func (c *webhookProcessor) GetWebhookProcessor(ctx context.Context, headers http.Header, body []byte) (domain.WebhookProcessor, error) {
	for _, processor := range c.processors {
		if processor.CanProcess(ctx, headers, body) {
			return processor, nil
		}
	}

	return nil, errors.New("could not find suitable WebhookProcessor")
}

func (c *webhookProcessor) StartHandleDeployConfigEvents(ctx context.Context) {
	logger := logging.FromContext(ctx)
	chans := make([]<-chan domain.DeployConfig, 0)
	for _, p := range c.processors {
		chans = append(chans, p.GetDeployConfigEvents(ctx))
	}
	out := getDeployConfigsChan(chans...)

	for {
		select {
		case cfg := <-out:
			logger.WithFields(log.Fields{
				"release":       cfg.ReleaseName,
				"chart_name":    cfg.ChartName,
				"chart_version": cfg.ChartVersion,
			}).Info("updating release")
			if err := c.helmService.DeployChart(ctx, cfg); err != nil {
				logger.WithFields(log.Fields{
					"chart_name": cfg.ChartName,
					"error":      cfg.ChartVersion,
				}).Error("could not deploy chart")
				continue
			}
			logger.WithField("release", cfg.ReleaseName).Debug("release updated")
		}
	}
}

func getDeployConfigsChan(chans ...<-chan domain.DeployConfig) <-chan domain.DeployConfig {
	out := make(chan domain.DeployConfig)
	go func() {
		var wg sync.WaitGroup
		wg.Add(len(chans))

		for _, c := range chans {
			go func(c <-chan domain.DeployConfig) {
				for v := range c {
					out <- v
				}
				wg.Done()
			}(c)
		}

		wg.Wait()
		close(out)
	}()
	return out
}
