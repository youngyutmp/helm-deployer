package service

import (
	"bytes"
	"time"

	"github.com/entwico/helm-deployer/domain"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

//ChartValuesServiceImpl is an implementation of ChartValuesService interface
type ChartValuesServiceImpl struct {
	Repository domain.ChartValuesRepository
}

//NewChartValuesService returns new instance of ChartValuesService
func NewChartValuesService(repository domain.ChartValuesRepository) domain.ChartValuesService {
	return &ChartValuesServiceImpl{
		Repository: repository,
	}
}

//FindAll returns all ChartValues objects
func (c *ChartValuesServiceImpl) FindAll() ([]domain.ChartValues, error) {
	return c.Repository.FindAll()
}

//FindOne returns ChartValues object by its id
func (c *ChartValuesServiceImpl) FindOne(id string) (*domain.ChartValues, error) {
	return c.Repository.FindOne(id)
}

//Create creates new ChartValues object
func (c *ChartValuesServiceImpl) Create(item *domain.ChartValues) (*domain.ChartValues, error) {
	item.ID = ""
	return c.Repository.Save(item)
}

//Update updates existing ChartValues object
func (c *ChartValuesServiceImpl) Update(id string, newItem *domain.ChartValues) (*domain.ChartValues, error) {
	item, err := c.Repository.FindOne(id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, errors.New("item not found")
	}
	if err := yaml.Unmarshal([]byte(newItem.Data), new(bytes.Buffer)); err != nil {
		return nil, errors.New("Data is not a valid YAML")
	}

	item.Name = newItem.Name
	item.Description = newItem.Description
	item.ChartName = newItem.ChartName
	item.Data = newItem.Data
	item.UpdatedAt = time.Now()
	return c.Repository.Save(item)
}

//Delete deletes ChartValues object
func (c *ChartValuesServiceImpl) Delete(id string) error {
	return c.Repository.Delete(id)
}
