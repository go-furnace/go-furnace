package db

import (
  "database/sql"
  "log"
  "path/filepath"

  // It's not directly used.
  _ "github.com/mattn/go-sqlite3"
  "github.com/Skarlso/go_aws_mine/utils"
  "github.com/Skarlso/go_aws_mine/cfg"
)

var configPath string

func init() {
  configPath = cfg.ConfigPath()
}


// InitDb initializes the database.
func InitDb() {
  db, err := sql.Open("sqlite3", filepath.Join(configPath, "go_aws_main.db"))
  if err != nil {
    log.Fatal(err)
  }
  defer db.Close()

  // Check if table already exists.

  res, err := db.Query("SELECT name FROM sqlite_master WHERE type='table' AND name='instances';")
  utils.CheckError(err)
  defer res.Close()
  if res.Next() {
    log.Println("Database already exists. Nothing to do.")
    return
  }

  sqlStmt := databaseSQL()
  _, err = db.Exec(sqlStmt)
  if err != nil {
    log.Fatalf("%q: %s\n", err, sqlStmt)
  }
  log.Println("Database created.")
}

func databaseSQL() string {
	return `create table instances (
    ip varchar(100),
    id varchar(100),
    PRIMARY KEY (id)
);`
}
