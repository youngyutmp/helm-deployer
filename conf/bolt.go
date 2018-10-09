package conf

import (
	"log"

	"github.com/boltdb/bolt"
	"github.com/xlab/closer"
)

// BoltConnect opens bolt database
func BoltConnect(config *Config) (*bolt.DB, error) {
	logger := config.LogConfig.Logger
	db, err := bolt.Open(config.DB.Path, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	closer.Bind(func() {
		logger.Info("closing database file")
		if err := db.Close(); err != nil {
			logger.WithField("error", err).Warn("could not close database file")
		}
	})

	return db, nil
}
