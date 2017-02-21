package plugins

import "log"

// MyAwesomePreCreatePlugin is a sample pre-create plugin.
type MyAwesomePreCreatePlugin struct {
	Name string
}

// RunPlugin is running a plugin.
func (mapcp MyAwesomePreCreatePlugin) RunPlugin() {
	log.Println("Awesome pre-create plugin event.")
}
