package db

import (
  "database/sql"
  "log"
  "path/filepath"
  "os/user"

  // Don't need to name it.
  _ "github.com/mattn/go-sqlite3"
  "github.com/Skarlso/go_aws_mine/utils"
)

var configPath string

func init() {
  // Get configuration path
  usr, err := user.Current()
  utils.CheckError(err)
  configPath = filepath.Join(usr.HomeDir, ".config", "go_aws_mine")
}


// InitDb initializes the database.
func InitDb() {
  db, err := sql.Open("sqlite3", filepath.Join(configPath, "go_aws_main.db"))
  if err != nil {
    log.Fatal(err)
  }
  defer db.Close()

  sqlStmt := databaseSQL()
  _, err = db.Exec(sqlStmt)
  if err != nil {
    log.Fatalf("%q: %s\n", err, sqlStmt)
  }
}

func databaseSQL() string {
	return `create table instances (
    ip varchar(100),
    id varchar(100),
    PRIMARY KEY (id)
);`
}
