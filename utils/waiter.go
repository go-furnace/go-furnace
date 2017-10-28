package utils

import (
	"fmt"
	"log"
	"sync"
	"time"

	config "github.com/Skarlso/go-furnace/config/common"
	googleconfig "github.com/Skarlso/go-furnace/config/google"
	"github.com/fatih/color"
	dm "google.golang.org/api/deploymentmanager/v2"
)

var yellow = color.New(color.FgYellow).SprintFunc()
var red = color.New(color.FgRed).SprintFunc()
var keyName = color.New(color.FgWhite, color.Bold).SprintFunc()

// WaitForFunctionWithStatusOutput waits for a function to complete its action.
func WaitForFunctionWithStatusOutput(state string, freq int, f func()) {
	var wg sync.WaitGroup
	wg.Add(1)
	done := make(chan bool)
	go func() {
		defer wg.Done()
		f()
		done <- true
	}()
	go func() {
		counter := 0
		for {
			counter = (counter + 1) % len(Spinners[config.SPINNER])
			fmt.Printf("\r[%s] Waiting for state: %s", yellow(string(Spinners[config.SPINNER][counter])), red(state))
			time.Sleep(time.Duration(freq) * time.Second)
			select {
			case <-done:
				fmt.Println()
				break
			default:
			}
		}
	}()

	wg.Wait()
}

// WaitForDeploymentToFinish waits for a google deployment to finish.
func WaitForDeploymentToFinish(d dm.Service, deploymentName string) {
	project := d.Deployments.Get(googleconfig.GOOGLEPROJECTNAME, deploymentName)
	deploymentOp, err := project.Do()
	if err != nil {
		log.Fatal("error while getting deployment: ", err)
	}
	var counter int
	// This needs a timeout
	for deploymentOp.Operation.Status == "RUNNING" {
		time.Sleep(1 * time.Duration(time.Second))
		counter = (counter + 1) % len(Spinners[config.SPINNER])
		fmt.Printf("\r[%s] Waiting for state: %s", yellow(string(Spinners[config.SPINNER][counter])), red("DONE"))
		deploymentOp, err = project.Do()
		if err != nil {
			log.Fatal("\nerror while getting deployment: ", err)
		}
	}
	fmt.Println()
	log.Println("Final deployment status: ", keyName(deploymentOp.Operation.Status))
}
