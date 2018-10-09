package domain

import (
	"context"
	"time"

	"net/http"

	"github.com/globalsign/mgo/bson"
)

//Webhook defines webhook structure
type Webhook struct {
	ID           bson.ObjectId          `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description,omitempty"`
	Condition    GitlabWebhookCondition `json:"condition"`
	DeployConfig DeployConfig           `json:"deployConfig"`
	CreatedAt    time.Time              `json:"createdAt"`
	UpdatedAt    time.Time              `json:"updatedAt"`
}

//GitlabWebhookCondition defines webhook condition structure
type GitlabWebhookCondition struct {
	WebhookType      string `json:"webhookType"`
	ProjectName      string `json:"projectName"`
	ProjectNamespace string `json:"projectNamespace"`
	GitRef           string `json:"gitRef"`
	IsTag            bool   `json:"isTag"`
}

//DeployConfig defines deploy config structure
type DeployConfig struct {
	ReleaseName   string  `json:"releaseName"`
	ChartName     string  `json:"chartName"`
	ChartVersion  string  `json:"chartVersion"`
	ChartValuesID *string `json:"chartValuesId"`
}

//WebhookService manages WebHooks
type WebhookService interface {
	FindAll() ([]Webhook, error)
	FindOne(id string) (*Webhook, error)
	Create(item *Webhook) (*Webhook, error)
	Update(id string, newItem *Webhook) (*Webhook, error)
	Delete(id string) error
}

//WebhookRepository persists Webhooks to the database
type WebhookRepository interface {
	FindAll() ([]Webhook, error)
	FindOne(id string) (*Webhook, error)
	FindByName(name string) (*Webhook, error)
	Save(item *Webhook) (*Webhook, error)
	Delete(id string) error
}

//WebhookDispatcher handles webhooks
type WebhookDispatcher interface {
	GetWebhookProcessor(ctx context.Context, headers http.Header, body []byte) (WebhookProcessor, error)
	StartHandleDeployConfigEvents(ctx context.Context)
}

//WebhookProcessor listens to Webhooks and deploys charts
type WebhookProcessor interface {
	//DetermineWebhookType(headers http.Header) (enums.WebhookType, error)
	//DeployChart(cfg DeployConfig) error
	CanProcess(ctx context.Context, headers http.Header, body []byte) bool
	Process(ctx context.Context, headers http.Header, body []byte) error
	GetDeployConfigEvents(ctx context.Context) chan DeployConfig
}
