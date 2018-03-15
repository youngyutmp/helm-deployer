package service

import (
	"fmt"

	"bytes"
	"encoding/json"
	"github.com/Sirupsen/logrus"
)

const (
	gitlabEventPush         = "Push Hook"
	gitlabEventTag          = "Tag Push Hook"
	gitlabEventIssue        = "Issue Hook"
	gitlabEventComment      = "Note Hook"
	gitlabEventMergeRequest = "Merge Request Hook"
	gitlabEventWikiPage     = "Wiki Page Hook"
	gitlabEventPipeline     = "Pipeline Hook"
	gitlabEventBuild        = "Build Hook"
)

type WebhookCallbackService interface {
	ProcessWebhook(webhookType string, webhookBody []byte) error
}

func NewGitlabWebhookCallbackService(webhookSvc WebhookService, chartRepoSvc ChartRepositoryService, chartValuesSvc ChartValuesService, helmSvc HelmService) WebhookCallbackService {
	return &GitlabWebhookCallbackService{
		WebhookService:         webhookSvc,
		ChartRepositoryService: chartRepoSvc,
		ChartValuesService:     chartValuesSvc,
		HelmService:            helmSvc,
	}
}

type GitlabWebhookCallbackService struct {
	WebhookService         WebhookService
	ChartRepositoryService ChartRepositoryService
	ChartValuesService     ChartValuesService
	HelmService            HelmService
}

func (c *GitlabWebhookCallbackService) ProcessWebhook(webhookType string, webhookBody []byte) error {
	switch webhookType {
	case gitlabEventPipeline:
		logrus.Debug("Processing pipeline hook")
		payload := new(WebhookGitlabPipeline)
		if err := json.Unmarshal(webhookBody, &payload); err != nil {
			return err
		}

		if payload.ObjectAttributes.Status == "success" {
			cond := WebhookCondition{
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

func (c *GitlabWebhookCallbackService) processCondition(cond WebhookCondition) error {
	dc, err := c.getDeployConfigs(cond)
	if err != nil {
		return err
	}
	for _, cfg := range dc {
		logrus.Debugf("Updating release %s. Chart %s %s", cfg.ReleaseName, cfg.ChartName, cfg.ChartVersion)
		err = c.deployChart(cfg)
		if err != nil {
			return err
		}
		logrus.Infof("Release %s updated. Chart %s %s", cfg.ReleaseName, cfg.ChartName, cfg.ChartVersion)
	}
	return nil
}

func (c *GitlabWebhookCallbackService) getDeployConfigs(cond WebhookCondition) ([]DeployConfig, error) {
	webhooks, err := c.WebhookService.FindAll()
	if err != nil {
		return nil, err
	}
	var dc []DeployConfig
	for _, w := range webhooks {
		if w.Condition == cond {
			dc = append(dc, w.DeployConfig)
		}
	}

	return dc, nil
}

func (c *GitlabWebhookCallbackService) deployChart(cfg DeployConfig) error {
	logrus.Debugf("Deploying chart %s %s", cfg.ChartName, cfg.ChartVersion)

	var rawVals []byte
	if cfg.ChartValuesId != nil {
		values, err := c.ChartValuesService.FindOne(*cfg.ChartValuesId)
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
