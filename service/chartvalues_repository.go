package service

import (
	"encoding/json"
	"time"

	"github.com/boltdb/bolt"
	"github.com/globalsign/mgo/bson"
	"errors"
)

var chartBucket = []byte("chartValues")

type ChartValuesRepository interface {
	FindAll() ([]ChartValues, error)
	FindOne(id string) (*ChartValues, error)
	FindByName(name string) (*ChartValues, error)
	Save(item *ChartValues) (*ChartValues, error)
	Delete(id string) error
}

type BoltChartValuesRepository struct {
	db *bolt.DB
}

func NewChartValuesRepository(db *bolt.DB) *BoltChartValuesRepository {
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(chartBucket)
		return err
	})
	return &BoltChartValuesRepository{db}
}

func (r *BoltChartValuesRepository) FindAll() ([]ChartValues, error) {
	items := []ChartValues{}

	err := r.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(chartBucket)
		b.ForEach(func(k, v []byte) error {
			item, err := decodeChartValues(v)
			if err != nil {
				return err
			}
			items = append(items, *item)
			return nil
		})

		return nil
	})
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (r *BoltChartValuesRepository) FindOne(id string) (*ChartValues, error) {
	var item *ChartValues
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

func (r *BoltChartValuesRepository) FindByName(name string) (*ChartValues, error) {
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

func (r *BoltChartValuesRepository) Save(item *ChartValues) (*ChartValues, error) {
	err := r.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(chartBucket)
		if item.ID == "" {
			item.ID = bson.NewObjectId()
			item.CreatedAt = time.Now()
			item.UpdatedAt = time.Now()
			enc, err := item.encodeChartValues()
			if err != nil {
				return err
			}
			return b.Put([]byte(item.ID.Hex()), enc)
		} else {
			enc, err := item.encodeChartValues()
			if err != nil {
				return err
			}
			return b.Put([]byte(item.ID.Hex()), enc)
		}
	})
	if err != nil {
		return nil, err
	}

	return item, nil
}

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

func (p *ChartValues) encodeChartValues() ([]byte, error) {
	enc, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return enc, nil
}

func decodeChartValues(data []byte) (*ChartValues, error) {
	var item *ChartValues
	err := json.Unmarshal(data, &item)
	if err != nil {
		return nil, err
	}
	return item, nil
}
