package conf

import (
	"log"

	"github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"github.com/xlab/closer"
)

// BoltConnect opens bolt database
func BoltConnect(config *Config) (*bolt.DB, error) {

	db, err := bolt.Open(config.DB.Path, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	closer.Bind(func() {
		logrus.Info("Closing database file")
		if err := db.Close(); err != nil {
			logrus.Warnf("Can't close database file: %v", err)
		}
	})

	return db, nil
}
