## DigitalOcean

Furnace now supports DigitalOcean. This is really through a library called [Yogsothoth](https://github.com/Skarlso/yogsothoth). Yogsothoth aims to provide the same experience to DigitalOcean assets that does CloudFormation for AWS services.

This means, that there is a configuration template that describes a set of resources bundled together beneath an umbrella called stack.

This library is in Alpha and only supports Droplets for now. Slowly more resources and features will be available much like the templates of CF and GCP.

### Commands

For now, only `create` is done. This will be improved rapidly as more functionality is available through Yogsothoth.