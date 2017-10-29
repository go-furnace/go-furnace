package awscommands

import (
	"fmt"
	"sync"
	"time"

	config "github.com/Skarlso/go-furnace/config/common"
	"github.com/fatih/color"
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
			counter = (counter + 1) % len(config.Spinners[config.SPINNER])
			fmt.Printf("\r[%s] Waiting for state: %s", yellow(string(config.Spinners[config.SPINNER][counter])), red(state))
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
