package plugins

import "log"

// MyAwesomePostCreatePlugin is a sample pre-create plugin.
type MyAwesomePostCreatePlugin struct {
	Name string
}

// RunPlugin is running a plugin.
func (mapcp MyAwesomePostCreatePlugin) RunPlugin() {
	log.Println("Awesome post-create plugin event.")
}
