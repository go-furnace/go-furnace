package commands

import (
  "os"
  "os/user"
  "path/filepath"

  "github.com/Yitsushi/go-commander"
  "github.com/Skarlso/go_aws_mine/utils"
)

// Init is an init command
type Init struct {}

// Execute initializes everything.
func (i *Init) Execute(opts *commander.CommandHelper) {
  usr, err := user.Current()
  utils.CheckError(err)

  err = os.Mkdir(filepath.Join(usr.HomeDir, ".config", "go_aws_mine"), os.ModeDir)
  utils.CheckError(err)

  // Copy files from cfg folder to ./config/go_aws_mine.
}
