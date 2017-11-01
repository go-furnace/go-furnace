package googlecloud

import (
	"fmt"
	"log"
	"time"

	config "github.com/Skarlso/go-furnace/config/common"
	googleconfig "github.com/Skarlso/go-furnace/config/google"
	"github.com/fatih/color"
	dm "google.golang.org/api/deploymentmanager/v2"
	"google.golang.org/api/googleapi"
)

var keyName = color.New(color.FgWhite, color.Bold).SprintFunc()
var yellow = color.New(color.FgYellow).SprintFunc()
var red = color.New(color.FgRed).SprintFunc()

// WaitForDeploymentToFinish waits for a google deployment to finish.
func WaitForDeploymentToFinish(d dm.Service, deploymentName string) {
	project := d.Deployments.Get(googleconfig.GOOGLEPROJECTNAME, deploymentName)
	deploymentOp, err := project.Do()
	if err != nil {
		config.HandleFatal("error while getting deployment: ", err)
	}
	var counter int
	// This needs a timeout
	for deploymentOp.Operation.Status == "RUNNING" || deploymentOp.Operation.Status == "PENDING" {
		deploymentOp, err = project.Do()
		if err != nil {
			if err.(*googleapi.Error).Code != 404 {
				config.HandleFatal("error while getting deployment: ", err)
			} else {
				log.Println("\nStack terminated!")
				break
			}
		}
		time.Sleep(1 * time.Duration(time.Second))
		counter = (counter + 1) % len(config.Spinners[config.SPINNER])
		fmt.Printf("\r[%s] Waiting for state: %s", yellow(string(config.Spinners[config.SPINNER][counter])), red("DONE"))
	}
}
