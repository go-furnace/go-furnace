package handle

import "log"

// LogFatalf is used to define the fatal error handler function. In unit tests, this is used to
// mock out fatal errors so we can test for them.
var LogFatalf = log.Fatalf

// Error extracts the if err != nil check. If the given error is not nil it will call
// the defined fatal error handler function.
func Error(err error) {
	if err != nil {
		Fatal("Error occurred:", err)
	}
}

// Fatal is a wrapper for LogFatalf function. It's used to throw a Fatal in case it needs to.
func Fatal(s string, err error) {
	LogFatalf(s, err)
}
