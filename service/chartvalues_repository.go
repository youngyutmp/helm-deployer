package service

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/coreos/bbolt"
	"github.com/entwico/helm-deployer/domain"
	"github.com/globalsign/mgo/bson"
)

var chartBucket = []byte("chartValues")

//BoltChartValuesRepository is an implementation of ChartValuesRepository which uses BoltDB
type BoltChartValuesRepository struct {
	db *bolt.DB
}

//NewChartValuesRepository returns a new instance of ChartValuesRepository
func NewChartValuesRepository(db *bolt.DB) (domain.ChartValuesRepository, error) {
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(chartBucket)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &BoltChartValuesRepository{db}, nil
}

//FindAll returns all ChartValues objects
func (r *BoltChartValuesRepository) FindAll() ([]domain.ChartValues, error) {
	items := make([]domain.ChartValues, 0)

	err := r.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(chartBucket)
		err := b.ForEach(func(k, v []byte) error {
			item, err := decodeChartValues(v)
			if err != nil {
				return err
			}
			items = append(items, *item)
			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return items, nil
}

//FindOne returns ChartValues object by its id
func (r *BoltChartValuesRepository) FindOne(id string) (*domain.ChartValues, error) {
	var item *domain.ChartValues
	err := r.db.View(func(tx *bolt.Tx) error {
		var err error
		b := tx.Bucket(chartBucket)
		k := []byte(id)
		itemData := b.Get(k)
		if len(itemData) == 0 {
			return nil
		}
		item, err = decodeChartValues(itemData)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return item, nil
}

//FindByName returns ChartsValue object by chart name
func (r *BoltChartValuesRepository) FindByName(name string) (*domain.ChartValues, error) {
	items, err := r.FindAll()
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		if item.Name == name {
			return &item, nil
		}
	}
	return nil, nil
}

//Save saves ChartValues object to the database
func (r *BoltChartValuesRepository) Save(item *domain.ChartValues) (*domain.ChartValues, error) {
	err := r.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(chartBucket)
		if item.ID == "" {
			item.ID = bson.NewObjectId()
			item.CreatedAt = time.Now()
			item.UpdatedAt = time.Now()
			enc, err := encodeChartValues(item)
			if err != nil {
				return err
			}
			return b.Put([]byte(item.ID.Hex()), enc)
		}
		enc, err := encodeChartValues(item)
		if err != nil {
			return err
		}
		return b.Put([]byte(item.ID.Hex()), enc)
	})
	if err != nil {
		return nil, err
	}

	return item, nil
}

//Delete removes ChartValues object from the database
func (r *BoltChartValuesRepository) Delete(id string) error {
	item, err := r.FindOne(id)
	if err != nil {
		return err
	}
	if item == nil {
		return errors.New("not found")
	}
	return r.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(chartBucket)
		k := []byte(id)
		return b.Delete(k)
	})
}

func encodeChartValues(p *domain.ChartValues) ([]byte, error) {
	enc, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return enc, nil
}

func decodeChartValues(data []byte) (*domain.ChartValues, error) {
	var item *domain.ChartValues
	err := json.Unmarshal(data, &item)
	if err != nil {
		return nil, err
	}
	return item, nil
}
