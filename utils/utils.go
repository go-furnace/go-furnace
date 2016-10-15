package utils

import "log"

// CheckError handles errors.
func CheckError(err error) {
  if err != nil {
    log.Fatalf("Error occurred: %s", err.Error())
  }
}
