package commands

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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
	client *http.Client
	ctx    context.Context
}

// Execute runs the create command
func (u *Update) Execute(opts *commander.CommandHelper) {
	override := opts.Flag("y")
	configName := opts.Arg(0)
	if len(configName) > 0 {
		dir, _ := os.Getwd()
		if err := fc.LoadConfigFileIfExists(dir, configName); err != nil {
			handle.Fatal(configName, err)
		}
	}
	d := NewDeploymentService(u.ctx, u.client)
	err := update(d, fc.Config.Main.ProjectName, override)
	handle.Error(err)
}

func update(d DeploymentmanagerService, projectName string, override bool) error {
	log.Println("Creating Deployment update under project name: .", keyName(projectName))

	deploymentName := fc.Config.Gcp.StackName

	fingerPrint, err := getFingerPrintForDeployment(d, deploymentName)
	if err != nil {
		return errors.Wrap(err, "failed to get fingerprint for deployment")
	}

	targetConfiguration := constructTargetConfiguration()
	previewDeployments := dm.Deployment{
		Name:        deploymentName,
		Target:      &targetConfiguration,
		Fingerprint: fingerPrint,
	}
	updateCall := d.Deployments.Update(projectName, deploymentName, &previewDeployments)

	if v, err := doPreview(d, updateCall, override); err != nil {
		return errors.Wrap(err, "failed doing preview for deployment")
	} else if !v { // the preview was cancelled
		return nil
	}

	// Getting the new fingerprint of the deployment in preview state
	fingerPrint, err = getFingerPrintForDeployment(d, deploymentName)
	if err != nil {
		return errors.Wrap(err, "failed to get fingerprint for deployment")
	}

	// Construct a new update request to finalise the update process. Target configuration must be left out.
	updateDeployments := dm.Deployment{
		Name:        deploymentName,
		Fingerprint: fingerPrint,
	}
	updateCall = d.Deployments.Update(projectName, deploymentName, &updateDeployments)
	err = doUpdate(d, updateCall)
	if err != nil {
		return errors.Wrap(err, "error in update function while calling doUpdate")
	}
	log.Println("Update done. Bye.")
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

func doPreview(d DeploymentmanagerService, call *dm.DeploymentsUpdateCall, override bool) (bool, error) {
	// Setup update policies and initiate preview
	call.Preview(true)
	deploymentName := fc.Config.Gcp.StackName
	// Do the preview call
	_, err := call.Do()
	if err != nil {
		return false, errors.Wrap(err, "error in cancelOrDoUpdate Do call")
	}
	waitForDeploymentToFinish(d, fc.Config.Main.ProjectName, deploymentName)

	log.Println("Please review the preview data.")

	if !override {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Would you like to apply the changes? (y/N):")
		confirm, _ := reader.ReadString('\n')
		confirm = strings.TrimSuffix(confirm, "\n")
		if confirm != "y" {
			log.Println("Cancelling without applying change set.")
			fingerPrint, err := getFingerPrintForDeployment(d, deploymentName)
			if err != nil {
				return false, errors.Wrap(err, "failed to get fingerprint for cancelling")
			}
			cancelCall := d.Deployments.CancelPreview(fc.Config.Main.ProjectName, deploymentName, &dm.DeploymentsCancelPreviewRequest{
				Fingerprint: fingerPrint,
			})
			_, err = cancelCall.Do()
			return false, errors.Wrap(err, "cancel preview call")
		}
	}

	return true, nil
}

// doUpdate will finalise the update process by submitting the same update request
// but without preview and without the target configuration.
func doUpdate(d DeploymentmanagerService, call *dm.DeploymentsUpdateCall) error {
	// Setup update policies
	call.CreatePolicy(fc.Config.Gcp.CreatePolicy)
	call.DeletePolicy(fc.Config.Gcp.DeletePolicy)

	// Do the preview call
	op, err := call.Do()
	if err != nil {
		return errors.Wrap(err, "error in cancelOrDoUpdate Do call")
	}
	waitForDeploymentToFinish(d, fc.Config.Main.ProjectName, fc.Config.Gcp.StackName)

	b, err := op.MarshalJSON()
	if err != nil {
		return errors.Wrap(err, "error in cancelOrDoUpdate MarshalJSON")
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
	ctx := context.Background()
	client, err := google.DefaultClient(ctx, dm.NdevCloudmanScope)
	handle.Error(err)
	u := Update{client: client, ctx: ctx}
	return &commander.CommandWrapper{
		Handler: &u,
		Help: &commander.CommandDescriptor{
			Name:             "update",
			ShortDescription: "Update updates a Google Deployment",
			LongDescription:  `Using a pre-configured yaml file, update a collection of resources using Deployment Manager Service.`,
			Arguments:        "custom-config [-y]",
			Examples:         []string{"", "custom-config"},
		},
	}
}
