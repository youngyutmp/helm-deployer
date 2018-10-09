package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/entwico/helm-deployer/domain"
	log "github.com/sirupsen/logrus"
)

const (
	headerWebhookNexus     = "X-Nexus-Webhook-Delivery"
	headerWebhookNexusType = "X-Nexus-Webhook-Id"
	nexusEventTypeAsset    = "rm:repository:asset"
)

type nexusWebhookProcessor struct {
	releaseProvider domain.K8SReleaseProvider
	events          chan domain.DeployConfig
	logger          *log.Entry
}

//NewNexusProcessor returns new instance of Nexus webhook processor
func NewNexusProcessor(releaseProvider domain.K8SReleaseProvider, logger *log.Entry) domain.WebhookProcessor {
	return &nexusWebhookProcessor{
		releaseProvider: releaseProvider,
		events:          make(chan domain.DeployConfig),
		logger:          logger,
	}
}

//CanProcess returns true if webhook can be processed by this processor
func (p *nexusWebhookProcessor) CanProcess(ctx context.Context, headers http.Header, body []byte) bool {
	val := headers.Get(headerWebhookNexus)
	if val != "" {
		return true
	}
	return false
}

//Process handles webhook
func (p *nexusWebhookProcessor) Process(ctx context.Context, headers http.Header, body []byte) error {
	p.logger.Info("processing Nexus webhook")
	event := headers.Get(headerWebhookNexusType)

	switch event {
	case nexusEventTypeAsset:
		return p.processAssetEvent(body)
	default:
		p.logger.WithField("event", event).Debug("skipping event")
	}
	return nil
}

func (p *nexusWebhookProcessor) GetDeployConfigEvents(ctx context.Context) chan domain.DeployConfig {
	return p.events
}

func (p *nexusWebhookProcessor) processAssetEvent(body []byte) error {
	p.logger.Debug("processing asset event")
	payload := new(WebhookNexusAsset)
	if err := json.Unmarshal(body, &payload); err != nil {
		return err
	}
	if payload.Asset.Format != "docker" || !(payload.Action == "UPDATED" || payload.Action == "CREATED") {
		p.logger.WithFields(log.Fields{
			"action": payload.Action,
			"format": payload.Asset.Format,
		}).Debug("ignoring asset event")
		return nil
	}

	path, tag := payload.GetRepositoryPathAndTag()
	if tag != "" {
		imagePath := fmt.Sprintf("%s:%s", path, tag)
		p.logger.WithFields(log.Fields{
			"image":      imagePath,
			"repository": payload.RepositoryName,
		}).Debug("image updated in repository")
		deployConfigs, err := p.releaseProvider.GetDeployConfigsForImagePath(imagePath)
		if err != nil {
			return err
		}
		for _, cfg := range deployConfigs {
			p.events <- *cfg
		}
	}

	return nil
}

//WebhookNexusAsset defines webhook payload structure
type WebhookNexusAsset struct {
	Timestamp      string `json:"timestamp"`
	NodeID         string `json:"nodeId"`
	Initiator      string `json:"initiator"`
	RepositoryName string `json:"repositoryName"`
	Action         string `json:"action"`
	Asset          struct {
		ID     string `json:"id"`
		Format string `json:"format"`
		Name   string `json:"name"`
	} `json:"asset"`
}

//GetRepositoryPathAndTag returns path and tag for docker image
func (r WebhookNexusAsset) GetRepositoryPathAndTag() (path string, tag string) {
	pathFirstIndex := strings.Index(r.Asset.Name, "/")
	pathLastIndex := strings.Index(r.Asset.Name, "/manifests/")
	if pathFirstIndex != -1 && pathLastIndex != -1 {
		path = r.Asset.Name[pathFirstIndex:pathLastIndex]
	}

	if tagLastIndex := strings.LastIndex(r.Asset.Name, "/"); tagLastIndex != -1 {
		tag = r.Asset.Name[tagLastIndex+1:]
	}
	if strings.HasPrefix(tag, "sha256:") {
		tag = ""
	}

	return path, tag
}
