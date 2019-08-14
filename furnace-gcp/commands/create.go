package commands

import (
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Yitsushi/go-commander"
	fc "github.com/go-furnace/go-furnace/furnace-gcp/config"
	"github.com/go-furnace/go-furnace/furnace-gcp/plugins"
	"github.com/go-furnace/go-furnace/handle"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	dm "google.golang.org/api/deploymentmanager/v2"
	"google.golang.org/api/option"
	"gopkg.in/yaml.v1"
)

// Create commands for google Deployment Manager
type Create struct {
}

// DeploymentService defines a service which implement `Insert`. This method
// inserts a deployment into a GCP project.
type DeploymentService interface {
	Insert(project string, deployment *dm.Deployment) *dm.DeploymentsInsertCall
	Delete(project string, deployment string) *dm.DeploymentsDeleteCall
	Get(project string, deployment string) *dm.DeploymentsGetCall
}

// DeploymentmanagerService defines a struct that we can use to mock GCP's
// deploymentmanager/v2 API.
type DeploymentmanagerService struct {
	*dm.Service
	Deployments DeploymentService
}

// NewDeploymentService will return a deployment manager service that
// can be used as a mock for the GCP deployment manager.
func NewDeploymentService(client *http.Client) DeploymentmanagerService {
	ctx := context.Background()
	d, _ := dm.NewService(ctx, option.WithHTTPClient(client))

	return DeploymentmanagerService{
		Deployments: d.Deployments,
	}
}

// Execute runs the create command
func (c *Create) Execute(opts *commander.CommandHelper) {
	configName := opts.Arg(0)
	if len(configName) > 0 {
		dir, _ := os.Getwd()
		if err := fc.LoadConfigFileIfExists(dir, configName); err != nil {
			handle.Fatal(configName, err)
		}
	}
	log.Println("Creating Deployment under project name: .", keyName(fc.Config.Main.ProjectName))
	deploymentName := fc.Config.Gcp.StackName
	log.Println("Deployment name is: ", keyName(deploymentName))
	ctx := context.Background()
	client, err := google.DefaultClient(ctx, dm.NdevCloudmanScope)
	handle.Error(err)
	d := NewDeploymentService(client)
	deployments := constructDeployment(deploymentName)
	plugins.RunPreCreatePlugins(deploymentName)
	err = insertDeployments(d, deployments, deploymentName)
	plugins.RunPostCreatePlugins(deploymentName)
	handle.Error(err)
}

func insertDeployments(d DeploymentmanagerService, deployments *dm.Deployment, deploymentName string) error {
	ret := d.Deployments.Insert(fc.Config.Main.ProjectName, deployments)
	if ret == nil {
		return errors.New("return value was nil")
	}
	_, err := ret.Do()
	if err != nil {
		return err
	}
	waitForDeploymentToFinish(d, fc.Config.Main.ProjectName, deploymentName)
	return nil
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

func constructDeployment(deploymentName string) *dm.Deployment {
	gConfig := fc.LoadGoogleStackConfig()
	configFile := dm.ConfigFile{
		Content: string(gConfig),
	}
	targetConfiguration := dm.TargetConfiguration{
		Config: &configFile,
	}

	imps := Imports{}
	err := yaml.Unmarshal(gConfig, &imps)
	handle.Error(err)

	// Load templates and all .schema files that might accompany them.
	if len(imps.Paths) > 0 {
		log.Println("Found the following import files: ", imps.Paths)
		var imports []*dm.ImportFile
		for _, temp := range imps.Paths {
			templateContent := fc.LoadImportFileContent(temp.Path)
			name := filepath.Base(temp.Path)
			if len(temp.Name) > 0 {
				name = temp.Name
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
			Arguments:        "custom-config",
			Examples:         []string{"", "custom-config"},
		},
	}
}
