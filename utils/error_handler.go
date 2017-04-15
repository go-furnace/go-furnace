package utils

import "log"

// LogFatalf is the function to log a fatal error.
var LogFatalf = log.Fatalf

// CheckError handles errors.
func CheckError(err error) {
	if err != nil {
		HandleFatal("Error occurred:", err)
	}
}

// HandleFatal handler fatal errors in Furnace.
func HandleFatal(s string, err error) {
	LogFatalf(s, err)
}
