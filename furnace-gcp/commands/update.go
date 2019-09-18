package commands

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"google.golang.org/api/googleapi"

	"github.com/pkg/errors"

	"golang.org/x/oauth2/google"
	"gopkg.in/yaml.v1"

	"github.com/Yitsushi/go-commander"
	fc "github.com/go-furnace/go-furnace/furnace-gcp/config"
	"github.com/go-furnace/go-furnace/handle"
	dm "google.golang.org/api/deploymentmanager/v2"
)

// Update defines and update command struct.
type Update struct {
}

// Execute runs the create command
func (u *Update) Execute(opts *commander.CommandHelper) {
	configName := opts.Arg(0)
	if len(configName) > 0 {
		dir, _ := os.Getwd()
		if err := fc.LoadConfigFileIfExists(dir, configName); err != nil {
			handle.Fatal(configName, err)
		}
	}
	err := update(fc.Config.Main.ProjectName)
	handle.Error(err)
}

func update(projectName string) error {
	log.Println("Creating Deployment update under project name: .", keyName(projectName))

	deploymentName := fc.Config.Gcp.StackName
	ctx := context.Background()
	client, err := google.DefaultClient(ctx, dm.NdevCloudmanScope)
	if err != nil {
		return errors.Wrap(err, "error in updater function while creating google client")
	}
	d := NewDeploymentService(ctx, client)
	fingerPrint, err := getFingerPrintForDeployment(d, deploymentName)
	if err != nil {
		return errors.Wrap(err, "failed to get fingerprint for deployment")
	}

	targetConfiguration := constructTargetConfiguration()
	deployments := dm.Deployment{
		Name:        deploymentName,
		Target:      &targetConfiguration,
		Fingerprint: fingerPrint,
	}
	updateCall := d.Deployments.Update(projectName, deploymentName, &deployments)
	err = cancelOrInsertUpdate(updateCall)
	if err != nil {
		return errors.Wrap(err, "error in update function while calling cancelOrInsertUpdate")
	}
	return nil
}

func getFingerPrintForDeployment(d DeploymentmanagerService, deploymentName string) (string, error) {
	project := d.Deployments.Get(fc.Config.Main.ProjectName, deploymentName)
	p, err := project.Do()
	if err != nil {
		if err.(*googleapi.Error).Code != 404 {
			return "", errors.Wrap(err, "failed to get deployment")
		}
		return "", errors.Wrap(err, "Stack not found!")
	}
	return p.Fingerprint, nil
}

func cancelOrInsertUpdate(call *dm.DeploymentsUpdateCall) error {
	// TODO: Make this work as a preview.
	//call.Preview(true)
	op, err := call.Do()
	if err != nil {
		return errors.Wrap(err, "error in cancelOrInsertUpdate Do call")
	}
	b, err := op.MarshalJSON()
	if err != nil {
		return errors.Wrap(err, "error in cancelOrInsertUpdate MarshalJSON")
	}
	fmt.Println(string(b))
	return nil
}

func constructTargetConfiguration() dm.TargetConfiguration {
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
			templateFile := dm.ImportFile{Content: string(templateContent), Name: name}
			imports = append(imports, &templateFile)
			if ok, schema := fc.LoadSchemaForPath(temp.Path); ok {
				f := dm.ImportFile{Content: string(schema)}
				imports = append(imports, &f)
			}
		}
		targetConfiguration.Imports = imports
	}

	return targetConfiguration
}

// NewUpdate creates a new update command
func NewUpdate(appName string) *commander.CommandWrapper {
	return &commander.CommandWrapper{
		Handler: &Update{},
		Help: &commander.CommandDescriptor{
			Name:             "update",
			ShortDescription: "Update updates a Google Deployment",
			LongDescription:  `Using a pre-configured yaml file, update a collection of resources using Deployment Manager Service.`,
			Arguments:        "custom-config",
			Examples:         []string{"", "custom-config"},
		},
	}
}
