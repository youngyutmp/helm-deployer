package domain

import (
	"time"

	"github.com/globalsign/mgo/bson"
)

//ChartValues represents a ChartValue object
type ChartValues struct {
	ID          bson.ObjectId `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	ChartName   string        `json:"chartName"`
	Data        string        `json:"data"`
	CreatedAt   time.Time     `json:"createdAt"`
	UpdatedAt   time.Time     `json:"updatedAt"`
}

//ChartValuesService manages ChartValues
type ChartValuesService interface {
	FindAll() ([]ChartValues, error)
	FindOne(id string) (*ChartValues, error)
	Create(item *ChartValues) (*ChartValues, error)
	Update(id string, newItem *ChartValues) (*ChartValues, error)
	Delete(id string) error
}

//ChartValuesRepository persists ChartValues to the database
type ChartValuesRepository interface {
	FindAll() ([]ChartValues, error)
	FindOne(id string) (*ChartValues, error)
	FindByName(name string) (*ChartValues, error)
	Save(item *ChartValues) (*ChartValues, error)
	Delete(id string) error
}
