package commands

import (
	"fmt"
	"log"
	"time"

	"github.com/fatih/color"
	"github.com/go-furnace/go-furnace/config"
	"github.com/go-furnace/go-furnace/handle"
	"google.golang.org/api/googleapi"
)

var keyName = color.New(color.FgWhite, color.Bold).SprintFunc()
var yellow = color.New(color.FgYellow).SprintFunc()
var red = color.New(color.FgRed).SprintFunc()

// WaitForDeploymentToFinish waits for a google deployment to finish.
func waitForDeploymentToFinish(d DeploymentmanagerService, projectName string, deploymentName string) {
	project := d.Deployments.Get(projectName, deploymentName)
	deploymentOp, err := project.Do()
	fmt.Println(deploymentOp)
	if err != nil {
		handle.Fatal("error while getting deployment: ", err)
	}
	var counter int
	// This needs a timeout
	for deploymentOp.Operation.Status == "RUNNING" || deploymentOp.Operation.Status == "PENDING" {
		deploymentOp, err = project.Do()
		if err != nil {
			if err.(*googleapi.Error).Code != 404 {
				handle.Fatal("error while getting deployment: ", err)
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
