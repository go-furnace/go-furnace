package commands

import (
	"errors"
	"log"
	"path/filepath"

	config "github.com/Skarlso/go-furnace/config"
	fc "github.com/Skarlso/go-furnace/gcp/config"
	"github.com/Yitsushi/go-commander"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	dm "google.golang.org/api/deploymentmanager/v2"
	yaml "gopkg.in/yaml.v1"
)

// Create commands for google Deployment Manager
type Create struct {
}

// Execute runs the create command
func (c *Create) Execute(opts *commander.CommandHelper) {
	if _, ok := opts.Opts["config"]; ok {
		file := opts.Opts["config"]
		if len(file) < 1 {
			config.HandleFatal("Please provide a filename with option -c", errors.New("missing filename"))
		}
		fc.Config.LoadConfiguration(file)
	}
	log.Println("Creating Deployment under project name: .", keyName(fc.Config.Main.ProjectName))
	deploymentName := fc.Config.Gcp.StackName
	log.Println("Deployment name is: ", keyName(deploymentName))
	ctx := context.Background()
	client, err := google.DefaultClient(ctx, dm.NdevCloudmanScope)
	if err != nil {
		log.Fatalf(err.Error())
	}
	d, _ := dm.New(client)
	deployments := constructDeploymen(deploymentName)
	ret := d.Deployments.Insert(fc.Config.Main.ProjectName, deployments)
	_, err = ret.Do()
	config.CheckError(err)
	waitForDeploymentToFinish(*d, fc.Config.Main.ProjectName, deploymentName)
}

// Path contains all the jinja imports in the config.yml file.
type Path struct {
	Path string `yaml:"path"`
	Name string `yaml:"name,omitempty"`
}

// Imports is the high level representation of imports in the config.yml file.
type Imports struct {
	Paths []Path `yaml:"imports"`
}

func constructDeploymen(deploymentName string) *dm.Deployment {
	gConfig := fc.LoadGoogleStackConfig()
	configFile := dm.ConfigFile{
		Content: string(gConfig),
	}
	targetConfiguration := dm.TargetConfiguration{
		Config: &configFile,
	}

	imps := Imports{}
	err := yaml.Unmarshal(gConfig, &imps)
	config.CheckError(err)

	// Load templates and all .schema files that might accompany them.
	if len(imps.Paths) > 0 {
		log.Println("Found the following import files: ", imps.Paths)
		imports := []*dm.ImportFile{}
		for _, temp := range imps.Paths {
			templateContent := fc.LoadImportFileContent(temp.Path)
			var name string
			if len(temp.Name) > 0 {
				name = temp.Name
			} else {
				name = filepath.Base(temp.Path)
			}
			log.Println("Adding template name: ", name)
			templateFile := &dm.ImportFile{Content: string(templateContent), Name: name}
			imports = append(imports, templateFile)
			if ok, schema := fc.LoadSchemaForPath(temp.Path); ok {
				f := &dm.ImportFile{Content: string(schema)}
				imports = append(imports, f)
			}
		}
		targetConfiguration.Imports = imports
	}

	deployments := dm.Deployment{
		Name:   deploymentName,
		Target: &targetConfiguration,
	}
	return &deployments
}

// NewCreate Creates a new create command
func NewCreate(appName string) *commander.CommandWrapper {
	return &commander.CommandWrapper{
		Handler: &Create{},
		Help: &commander.CommandDescriptor{
			Name:             "create",
			ShortDescription: "Create a Google Deployment Manager",
			LongDescription:  `Using a pre-configured yaml file, create a collection of resources using Deployment Manager Service.`,
			Arguments:        "[--config=configFile]",
			Examples:         []string{"create [--config=configFile]"},
		},
	}
}
