package service

import (
	"encoding/json"
	"time"

	"github.com/boltdb/bolt"
	"github.com/globalsign/mgo/bson"
	"github.com/pkg/errors"
)

var webhooksBucket = []byte("webhooks")

type WebhookRepository interface {
	FindAll() ([]Webhook, error)
	FindOne(id string) (*Webhook, error)
	FindByName(name string) (*Webhook, error)
	Save(item *Webhook) (*Webhook, error)
	Delete(id string) error
}

type BoltWebhookRepository struct {
	db *bolt.DB
}

func NewWebhookRepository(db *bolt.DB) *BoltWebhookRepository {
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(webhooksBucket)
		return err
	})
	return &BoltWebhookRepository{db}
}

func (r *BoltWebhookRepository) FindAll() ([]Webhook, error) {
	items := []Webhook{}

	err := r.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(webhooksBucket)
		b.ForEach(func(k, v []byte) error {
			item, err := decodeWebhook(v)
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

func (r *BoltWebhookRepository) FindOne(id string) (*Webhook, error) {
	var item *Webhook
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

func (r *BoltWebhookRepository) FindByName(name string) (*Webhook, error) {
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

func (r *BoltWebhookRepository) Save(item *Webhook) (*Webhook, error) {
	err := r.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(webhooksBucket)
		if item.ID == "" {
			item.ID = bson.NewObjectId()
			item.CreatedAt = time.Now()
			item.UpdatedAt = time.Now()
			enc, err := item.encodeWebhook()
			if err != nil {
				return err
			}
			return b.Put([]byte(item.ID.Hex()), enc)
		} else {
			enc, err := item.encodeWebhook()
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

func (p *Webhook) encodeWebhook() ([]byte, error) {
	enc, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return enc, nil
}

func decodeWebhook(data []byte) (*Webhook, error) {
	var item *Webhook
	err := json.Unmarshal(data, &item)
	if err != nil {
		return nil, err
	}
	return item, nil
}
