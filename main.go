package main

import (
	"drill/fetcher"
	"drill/mock"
	"drill/models"
	"drill/ui"
	"flag"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

func main() {
	// Parse command line flags
	mockMode := flag.Bool("mock", false, "Run in mock mode with simulated data")
	aggregateID := flag.String("id", "", "Aggregate ID (UUID) to query")
	servicesStr := flag.String("services", "", "Comma-separated list of service URLs (e.g., http://localhost:8081,http://localhost:8082)")
	flag.Parse()

	// Validate aggregate ID
	if *aggregateID == "" {
		if *mockMode {
			// Generate a random UUID for mock mode if not provided
			*aggregateID = uuid.New().String()
		} else {
			fmt.Println("Error: aggregate ID is required. Use -id flag.")
			fmt.Println("\nUsage:")
			flag.PrintDefaults()
			os.Exit(1)
		}
	}

	// Validate UUID format
	if _, err := uuid.Parse(*aggregateID); err != nil {
		fmt.Printf("Error: invalid UUID format for aggregate ID: %s\n", *aggregateID)
		os.Exit(1)
	}

	// Create the TUI model
	model := ui.NewModel(*aggregateID)

	// Create program
	p := tea.NewProgram(model, tea.WithAltScreen())

	// Start data loading in goroutine
	go func() {
		var events []models.Event
		var commands []models.Command
		var err error

		if *mockMode {
			// Use mock data
			events, commands = mock.GenerateMockData(*aggregateID)
		} else {
			// Parse services
			if *servicesStr == "" {
				p.Send(ui.ErrorMsg{Err: fmt.Errorf("no services specified. Use -services flag or -mock mode")})
				return
			}

			serviceURLs := strings.Split(*servicesStr, ",")
			var services []models.ServiceConfig
			for i, url := range serviceURLs {
				url = strings.TrimSpace(url)
				if url == "" {
					continue
				}
				services = append(services, models.ServiceConfig{
					Name: fmt.Sprintf("service-%d", i+1),
					URL:  url,
				})
			}

			// Fetch data from all services
			f := fetcher.NewFetcher(services)
			events, commands, err = f.FetchAll(*aggregateID)
			if err != nil {
				p.Send(ui.ErrorMsg{Err: err})
				return
			}
		}

		p.Send(ui.DataLoadedMsg{
			Events:   events,
			Commands: commands,
		})
	}()

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
