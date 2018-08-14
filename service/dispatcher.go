package service

import (
	"net/http"

	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/entwico/helm-deployer/domain"
	"github.com/pkg/errors"
)

type webhookProcessor struct {
	helmService domain.HelmService
	processors  []domain.WebhookProcessor
}

//NewWebhookDispatcher returns a new instance of WebhookProcessor
func NewWebhookDispatcher(helmService domain.HelmService, processors []domain.WebhookProcessor) domain.WebhookDispatcher {
	return &webhookProcessor{helmService: helmService, processors: processors}
}

func (c *webhookProcessor) GetWebhookProcessor(headers http.Header, body []byte) (domain.WebhookProcessor, error) {
	for _, processor := range c.processors {
		if processor.CanProcess(headers, body) {
			return processor, nil
		}
	}

	return nil, errors.New("could not find suitable WebhookProcessor")
}

func (c *webhookProcessor) StartHandleDeployConfigEvents() {
	chans := make([]<-chan domain.DeployConfig, 0)
	for _, p := range c.processors {
		chans = append(chans, p.GetDeployConfigEvents())
	}
	out := getDeployConfigsChan(chans...)

	for {
		select {
		case cfg := <-out:
			logrus.Debugf("updating release %s with chart %s %s", cfg.ReleaseName, cfg.ChartName, cfg.ChartVersion)
			if err := c.helmService.DeployChart(cfg); err != nil {
				logrus.Errorf("could not deploy chart %s: %v", cfg.ChartName, err)
				continue
			}
			logrus.Debugf("release %s updated", cfg.ReleaseName)
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
