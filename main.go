package main

import (
	"drill/models"
	"drill/ui"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Parse services from environment variable
	// Format: DRILL_SERVICES="service-name,idType,url;service-name2,idType2,url2"
	// Example: DRILL_SERVICES="account-service,indexId,https://account.com;payment-service,aggregateId,https://payment.com"
	services := parseServicesFromEnv()

	// Create the entry screen model
	model := ui.NewEntryModel(services)

	// Create and run program
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}

func parseServicesFromEnv() []models.ServiceConfig {
	envValue := os.Getenv("DRILL_SERVICES")
	if envValue == "" {
		return nil
	}

	var services []models.ServiceConfig

	// Split by semicolon for each service
	serviceStrs := strings.Split(envValue, ";")
	for _, svcStr := range serviceStrs {
		svcStr = strings.TrimSpace(svcStr)
		if svcStr == "" {
			continue
		}

		// Split by comma: name,idType,url
		parts := strings.Split(svcStr, ",")
		if len(parts) != 3 {
			fmt.Fprintf(os.Stderr, "Warning: invalid service format '%s', expected 'name,idType,url'\n", svcStr)
			continue
		}

		name := strings.TrimSpace(parts[0])
		idTypeStr := strings.TrimSpace(parts[1])
		url := strings.TrimSpace(parts[2])

		var idType models.IDType
		switch strings.ToLower(idTypeStr) {
		case "indexid":
			idType = models.IDTypeIndex
		case "aggregateid":
			idType = models.IDTypeAggregate
		default:
			fmt.Fprintf(os.Stderr, "Warning: invalid idType '%s' for service '%s', using aggregateId\n", idTypeStr, name)
			idType = models.IDTypeAggregate
		}

		services = append(services, models.ServiceConfig{
			Name:   name,
			IDType: idType,
			URL:    url,
		})
	}

	return services
}
