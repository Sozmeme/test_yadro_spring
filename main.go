package main

import (
	"fmt"
	"os"
	"yadro_test/utils"
)

func main() {
	configPath := "sunny_5_skiers/config.json"
	eventsPath := "sunny_5_skiers/events"

	configData, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Printf("Error reading config file: %v\n", err)
		return
	}

	eventData, err := os.ReadFile(eventsPath)
	if err != nil {
		fmt.Printf("Error reading events file: %v\n", err)
		return
	}

	processor, err := utils.NewProcessor(string(configData), string(eventData))
	if err != nil {
		fmt.Printf("Error creating processor: %v\n", err)
		return
	}

	logOutput := processor.ProcessEvents()

	summary := processor.GenerateSummary()

	fmt.Println("=== Output log ===")
	fmt.Println(logOutput)

	fmt.Println("\n=== Resulting table ===")
	fmt.Println(summary)
}
