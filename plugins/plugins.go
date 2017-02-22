// Package plugins will be a rudementary plugin functionality. Once the plugins for
// Go are finalized, I will use that way for better plugin support.
package plugins

// Plugin interface defines the capabilities of a plugin
type Plugin interface {
	RunPlugin()
}

// Plugins is a map of plugins for various stages of a deployment.
// pre-create: Plugins which will run before the creation of a stack.
// post-create: Plugins which will be called after the creation of a stack.
// pre-destroy: Plugins which will be called before a stack is destroyed.
// post-destroy: Plugins which will be called after a stack is destroyed.
var plugins map[string][]Plugin

func init() {
	plugins = make(map[string][]Plugin)
}

// RegisterPlugin registers a plugin for a given event.
func RegisterPlugin(event string, plugs []Plugin) {
	if v, ok := plugins[event]; ok {
		v = append(v, plugs...)
		plugins[event] = v
	} else {
		plugins[event] = plugs
	}
}

// GetPluginsForEvent returns all the registered plugins for event.
func GetPluginsForEvent(event string) []Plugin {
	return plugins[event]
}
