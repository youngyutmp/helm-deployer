package service

import (
	"encoding/json"
	"time"

	"github.com/boltdb/bolt"
	"github.com/entwico/helm-deployer/domain"
	"github.com/globalsign/mgo/bson"
	"github.com/pkg/errors"
)

var webhooksBucket = []byte("webhooks")

//BoltWebhookRepository is an implementation of the WebhookRepository interface which uses BoldDB
type BoltWebhookRepository struct {
	db *bolt.DB
}

//NewWebhookRepository returns a new instance of the WebhookRepository
func NewWebhookRepository(db *bolt.DB) (domain.WebhookRepository, error) {
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(webhooksBucket)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &BoltWebhookRepository{db}, nil
}

//FindAll returns all Webhooks
func (r *BoltWebhookRepository) FindAll() ([]domain.Webhook, error) {
	items := make([]domain.Webhook, 0)

	err := r.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(webhooksBucket)
		err := b.ForEach(func(k, v []byte) error {
			item, err := decodeWebhook(v)
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

//FindOne returns a Webhook by its id
func (r *BoltWebhookRepository) FindOne(id string) (*domain.Webhook, error) {
	var item *domain.Webhook
	err := r.db.View(func(tx *bolt.Tx) error {
		var err error
		b := tx.Bucket(webhooksBucket)
		k := []byte(id)
		itemData := b.Get(k)
		if len(itemData) == 0 {
			return nil
		}
		item, err = decodeWebhook(itemData)
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

//FindByName returns a Webhook by name
func (r *BoltWebhookRepository) FindByName(name string) (*domain.Webhook, error) {
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

//Save persists Webhook the the database
func (r *BoltWebhookRepository) Save(item *domain.Webhook) (*domain.Webhook, error) {
	err := r.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(webhooksBucket)
		if item.ID == "" {
			item.ID = bson.NewObjectId()
			item.CreatedAt = time.Now()
			item.UpdatedAt = time.Now()
			enc, err := encodeWebhook(item)
			if err != nil {
				return err
			}
			return b.Put([]byte(item.ID.Hex()), enc)
		}
		enc, err := encodeWebhook(item)
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

//Delete removes a Webhook
func (r *BoltWebhookRepository) Delete(id string) error {
	item, err := r.FindOne(id)
	if err != nil {
		return err
	}
	if item == nil {
		return errors.New("not found")
	}
	return r.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(webhooksBucket)
		k := []byte(id)
		return b.Delete(k)
	})
}

func encodeWebhook(p *domain.Webhook) ([]byte, error) {
	enc, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return enc, nil
}

func decodeWebhook(data []byte) (*domain.Webhook, error) {
	var item *domain.Webhook
	err := json.Unmarshal(data, &item)
	if err != nil {
		return nil, err
	}
	return item, nil
}
