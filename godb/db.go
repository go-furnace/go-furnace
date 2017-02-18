package godb

import (
	"log"
	"path/filepath"

	"github.com/Skarlso/go-furnace/config"
	"github.com/Skarlso/go-furnace/errorhandler"
	"github.com/boltdb/bolt"
)

// BUCKET is the main Bucket name for boltdb.
const BUCKET = "instances"

var configPath string

func init() {
	configPath = config.Path()
}

// InitDb initializes the database.
func InitDb() {
	db, err := bolt.Open(filepath.Join(configPath, "furnace_main.db"), 0600, nil)
	errorhandler.CheckError(err)
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		if _, dberr := tx.CreateBucketIfNotExists([]byte(BUCKET)); err != nil {
			return dberr
		}
		return nil
	})
	errorhandler.CheckError(err)

	log.Println("Database created.")
}
