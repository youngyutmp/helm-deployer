package service

import (
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/pkg/errors"
)

type Webhook struct {
	ID           bson.ObjectId    `json:"id"`
	Name         string           `json:"name"`
	Description  string           `json:"description,omitempty"`
	Condition    WebhookCondition `json:"condition"`
	DeployConfig DeployConfig     `json:"deployConfig"`
	CreatedAt    time.Time        `json:"createdAt"`
	UpdatedAt    time.Time        `json:"updatedAt"`
}

type WebhookCondition struct {
	WebhookType      string `json:"webhookType"`
	ProjectName      string `json:"projectName"`
	ProjectNamespace string `json:"projectNamespace"`
	GitRef           string `json:"gitRef"`
	IsTag            bool   `json:"isTag"`
}

type DeployConfig struct {
	ReleaseName   string  `json:"releaseName"`
	ChartName     string  `json:"chartName"`
	ChartVersion  string  `json:"chartVersion"`
	ChartValuesId *string `json:"chartValuesId"`
}

type WebhookService interface {
	FindAll() ([]Webhook, error)
	FindOne(id string) (*Webhook, error)
	Create(item *Webhook) (*Webhook, error)
	Update(id string, newItem *Webhook) (*Webhook, error)
	Delete(id string) error
}

type WebhookServiceImpl struct {
	Repository WebhookRepository
}

func NewWebhookService(repository WebhookRepository) *WebhookServiceImpl {
	return &WebhookServiceImpl{
		Repository: repository,
	}
}

func (c *WebhookServiceImpl) FindAll() ([]Webhook, error) {
	return c.Repository.FindAll()
}

func (c *WebhookServiceImpl) FindOne(id string) (*Webhook, error) {
	return c.Repository.FindOne(id)
}

func (c *WebhookServiceImpl) Create(item *Webhook) (*Webhook, error) {
	item.ID = ""
	return c.Repository.Save(item)
}

func (c *WebhookServiceImpl) Update(id string, newItem *Webhook) (*Webhook, error) {
	item, err := c.Repository.FindOne(id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, errors.New("item not found")
	}

	item.Name = newItem.Name
	item.Description = newItem.Description
	item.Condition = newItem.Condition
	item.DeployConfig = newItem.DeployConfig
	item.UpdatedAt = time.Now()
	return c.Repository.Save(item)
}

func (c *WebhookServiceImpl) Delete(id string) error {
	return c.Repository.Delete(id)
}
