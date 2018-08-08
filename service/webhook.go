package service

import (
	"time"

	"github.com/entwico/helm-deployer/domain"
	"github.com/pkg/errors"
)

//WebhookServiceImpl is an implementation of the WebhookService interface
type WebhookServiceImpl struct {
	Repository domain.WebhookRepository
}

//NewWebhookService returns a new instance of WebhookService
func NewWebhookService(repository domain.WebhookRepository) domain.WebhookService {
	return &WebhookServiceImpl{
		Repository: repository,
	}
}

//FindAll returns all Webhook objects
func (c *WebhookServiceImpl) FindAll() ([]domain.Webhook, error) {
	return c.Repository.FindAll()
}

//FindOne returns Webhook by its id
func (c *WebhookServiceImpl) FindOne(id string) (*domain.Webhook, error) {
	return c.Repository.FindOne(id)
}

//Create creates new Webhook
func (c *WebhookServiceImpl) Create(item *domain.Webhook) (*domain.Webhook, error) {
	item.ID = ""
	return c.Repository.Save(item)
}

//Update updates existing Webhook
func (c *WebhookServiceImpl) Update(id string, newItem *domain.Webhook) (*domain.Webhook, error) {
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

//Delete removes Webhook
func (c *WebhookServiceImpl) Delete(id string) error {
	return c.Repository.Delete(id)
}
