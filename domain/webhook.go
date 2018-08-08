package domain

import (
	"time"

	"github.com/globalsign/mgo/bson"
)

//Webhook defines webhook structure
type Webhook struct {
	ID           bson.ObjectId    `json:"id"`
	Name         string           `json:"name"`
	Description  string           `json:"description,omitempty"`
	Condition    WebhookCondition `json:"condition"`
	DeployConfig DeployConfig     `json:"deployConfig"`
	CreatedAt    time.Time        `json:"createdAt"`
	UpdatedAt    time.Time        `json:"updatedAt"`
}

//WebhookCondition defines webhook condition structure
type WebhookCondition struct {
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
