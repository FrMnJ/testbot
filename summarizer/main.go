package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/FrMnJ/testbot/pkg/testbot"
)

func main() {
	var scenarios []testbot.ResultScenario

	parts := []string{"results_part_1.json", "results_part_2.json", "results_part_3.json"}
	for _, part := range parts {
		scenariosPart, err := LoadResultScenarios(part)
		if err != nil {
			log.Printf("Error loading scenarios from %s: %v", part, err)
			continue
		}
		scenarios = append(scenarios, scenariosPart...)
	}

	totalSuccess := 0
	scoreTotal := 0.0
	for _, score := range scenarios {
		scoreTotal += float64(score.Score)
		if score.IsSuccess {
			totalSuccess++
		}
	}

	averageScore := scoreTotal / float64(len(scenarios))
	log.Printf("Average Score: %.2f", averageScore)

	succesRate := float64(totalSuccess) / float64(len(scenarios)) * 100
	log.Printf("Success Rate: %.2f%%", succesRate)
}

func LoadResultScenarios(part string) ([]testbot.ResultScenario, error) {
	file, err := os.ReadFile(part)
	if err != nil {
		log.Printf("Error reading scenarios file %s: %v", part, err)
		return nil, err
	}

	var scenarios []testbot.ResultScenario
	err = json.Unmarshal(file, &scenarios)
	if err != nil {
		log.Printf("Error unmarshalling scenarios from file %s: %v", part, err)
		return nil, err
	}
	return scenarios, nil
}
