package main

import (
	"fmt"
	"log"

	"github.com/shirou/gopsutil/v4/load"
)

func main() {
	// Test load.Misc() to verify process state counting
	misc, err := load.Misc()
	if err != nil {
		log.Fatalf("Error calling load.Misc(): %v", err)
	}

	fmt.Printf("Process State Counting Results:\n")
	fmt.Printf("==============================\n")
	fmt.Printf("ProcsTotal:   %d\n", misc.ProcsTotal)
	fmt.Printf("ProcsRunning: %d\n", misc.ProcsRunning)
	fmt.Printf("ProcsBlocked: %d\n", misc.ProcsBlocked)
	fmt.Printf("ProcsCreated: %d\n", misc.ProcsCreated)
	fmt.Printf("Ctxt:         %d\n", misc.Ctxt)
}
