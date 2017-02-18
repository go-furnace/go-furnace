package utils

import (
	"fmt"
	"sync"
	"time"
)

var spinner = 7

// WaitForFunctionWithStatusOutput waits for a function to complete its action.
func WaitForFunctionWithStatusOutput(state string, f func()) {
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
			counter = (counter + 1) % len(Spinners[spinner])
			fmt.Printf("\r\033[36m[%s]\033[m Waiting for stack to be in state: %s", string(Spinners[spinner][counter]), state)
			time.Sleep(1 * time.Second)
			select {
			case <-done:
				break
			default:
			}
		}
	}()

	wg.Wait()
}
