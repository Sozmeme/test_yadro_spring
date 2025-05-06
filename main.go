package main

import (
	"fmt"
	"os"
)


func main() {
	CONFIG_PATH := "sunny_5_skiers/config.json"
	EVENTS_PATH := "sunny_5_skiers/events"
	
	configData, err := os.ReadFile(CONFIG_PATH)
	if err != nil {
		fmt.Println("Ошибка чтения config.json:", err)
		return
	}

	eventData, err := os.ReadFile(EVENTS_PATH)
	if err != nil {
		fmt.Println("Ошибка чтения events.txt:", err)
		return
	}

	processor := createProcessor(string(configData), string(eventData))

	logOutput := processor.ProcessEvents()

	summary := processor.GenerateSummary()

	fmt.Println("=== Лог событий ===")
	fmt.Println(logOutput)

	fmt.Println("\n=== Финальный отчёт ===")
	fmt.Println(summary)
}
