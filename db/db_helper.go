package db

import (
	"log"
	"path/filepath"

	"github.com/Skarlso/go_aws_mine/config"
	"github.com/Skarlso/go_aws_mine/utils"
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
	db, err := bolt.Open(filepath.Join(configPath, "go_aws_main.db"), 0600, nil)
	utils.CheckError(err)
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte(BUCKET)); err != nil {
			return err
		}
		return nil
	})

	log.Println("Database created.")
}
