package service

import (
	"bytes"
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type ChartValues struct {
	ID          bson.ObjectId `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	ChartName   string        `json:"chartName"`
	Data        string        `json:"data"`
	CreatedAt   time.Time     `json:"createdAt"`
	UpdatedAt   time.Time     `json:"updatedAt"`
}

type ChartValuesService interface {
	FindAll() ([]ChartValues, error)
	FindOne(id string) (*ChartValues, error)
	Create(item *ChartValues) (*ChartValues, error)
	Update(id string, newItem *ChartValues) (*ChartValues, error)
	Delete(id string) error
}

type ChartValuesServiceImpl struct {
	Repository ChartValuesRepository
}

func NewChartValuesService(repository ChartValuesRepository) *ChartValuesServiceImpl {
	return &ChartValuesServiceImpl{
		Repository: repository,
	}
}

func (c *ChartValuesServiceImpl) FindAll() ([]ChartValues, error) {
	return c.Repository.FindAll()
}

func (c *ChartValuesServiceImpl) FindOne(id string) (*ChartValues, error) {
	return c.Repository.FindOne(id)
}

func (c *ChartValuesServiceImpl) Create(item *ChartValues) (*ChartValues, error) {
	item.ID = ""
	return c.Repository.Save(item)
}

func (c *ChartValuesServiceImpl) Update(id string, newItem *ChartValues) (*ChartValues, error) {
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

func (c *ChartValuesServiceImpl) Delete(id string) error {
	return c.Repository.Delete(id)
}
