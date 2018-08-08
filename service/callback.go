package service

import (
	"fmt"

	"bytes"
	"encoding/json"

	"github.com/Sirupsen/logrus"
	"github.com/entwico/helm-deployer/domain"
)

const (
	gitlabEventPipeline = "Pipeline Hook"
)

//GitlabWebhookCallbackService is an implementation of the WebhookCallbackService which reacts to gitlab webhooks
type GitlabWebhookCallbackService struct {
	WebhookService         domain.WebhookService
	ChartRepositoryService domain.ChartRepositoryService
	ChartValuesService     domain.ChartValuesService
	HelmService            domain.HelmService
}

//NewGitlabWebhookCallbackService returns a new instance of WebhookCallbackService
func NewGitlabWebhookCallbackService(webhookSvc domain.WebhookService, chartRepoSvc domain.ChartRepositoryService, chartValuesSvc domain.ChartValuesService, helmSvc domain.HelmService) domain.WebhookCallbackService {
	return &GitlabWebhookCallbackService{
		WebhookService:         webhookSvc,
		ChartRepositoryService: chartRepoSvc,
		ChartValuesService:     chartValuesSvc,
		HelmService:            helmSvc,
	}
}

//ProcessWebhook reacts to webhook
func (c *GitlabWebhookCallbackService) ProcessWebhook(webhookType string, webhookBody []byte) error {
	switch webhookType {
	case gitlabEventPipeline:
		logrus.Debug("Processing pipeline hook")
		payload := new(WebhookGitlabPipeline)
		if err := json.Unmarshal(webhookBody, &payload); err != nil {
			return err
		}

		if payload.ObjectAttributes.Status == "success" {
			cond := domain.WebhookCondition{
				WebhookType:      payload.ObjectKind,
				ProjectName:      payload.Project.Name,
				ProjectNamespace: payload.Project.Namespace,
				GitRef:           payload.ObjectAttributes.Ref,
				IsTag:            payload.ObjectAttributes.Tag,
			}

			if err := c.processCondition(cond); err != nil {
				logrus.Error(err)
				return err
			}
		} else {
			logrus.Infof(fmt.Sprintf("Skipping pipeline status '%s'", payload.ObjectAttributes.Status))
		}

	default:
		logrus.Error(fmt.Sprintf("Webhook '%s' is not supported", webhookType))
	}

	return nil
}

func (c *GitlabWebhookCallbackService) processCondition(cond domain.WebhookCondition) error {
	dc, err := c.getDeployConfigs(cond)
	if err != nil {
		return err
	}
	for _, cfg := range dc {
		logrus.Debugf("Updating release %s. Chart %s %s", cfg.ReleaseName, cfg.ChartName, cfg.ChartVersion)
		err = c.DeployChart(cfg)
		if err != nil {
			return err
		}
		logrus.Infof("Release %s updated. Chart %s %s", cfg.ReleaseName, cfg.ChartName, cfg.ChartVersion)
	}
	return nil
}

func (c *GitlabWebhookCallbackService) getDeployConfigs(cond domain.WebhookCondition) ([]domain.DeployConfig, error) {
	webhooks, err := c.WebhookService.FindAll()
	if err != nil {
		return nil, err
	}
	var dc []domain.DeployConfig
	for _, w := range webhooks {
		if w.Condition == cond {
			dc = append(dc, w.DeployConfig)
		}
	}

	return dc, nil
}

//DeployChart deploys the helm chart
func (c *GitlabWebhookCallbackService) DeployChart(cfg domain.DeployConfig) error {
	logrus.Debugf("Deploying chart %s %s", cfg.ChartName, cfg.ChartVersion)

	var rawVals []byte
	if cfg.ChartValuesID != nil {
		values, err := c.ChartValuesService.FindOne(*cfg.ChartValuesID)
		if err != nil {
			return err
		}
		if values != nil {
			rawVals = []byte(values.Data)
		}
	}

	data, err := c.ChartRepositoryService.GetChartData(cfg.ChartName, cfg.ChartVersion)
	if err != nil {
		return err
	}

	_, err = c.HelmService.UpdateRelease(cfg.ReleaseName, bytes.NewReader(data), rawVals)
	return err
}
