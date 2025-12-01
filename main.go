package main

import (
	"bufio"
	"drill/models"
	"drill/ui"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Parse services from CSV file
	services := parseServicesFromFile()

	// Create the entry screen model
	model := ui.NewEntryModel(services)

	// Create and run program
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}

func parseServicesFromFile() []models.ServiceConfig {
	// Look for .drill.csv in current directory, then home directory
	paths := []string{
		".drill.csv",
		filepath.Join(os.Getenv("HOME"), ".drill.csv"),
	}

	var file *os.File

	for _, path := range paths {
		f, err := os.Open(path)
		if err == nil {
			file = f
			break
		}
	}

	if file == nil {
		return nil
	}
	defer file.Close()

	var services []models.ServiceConfig
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split by comma: name,idType,url
		parts := strings.Split(line, ",")
		if len(parts) != 3 {
			fmt.Fprintf(os.Stderr, "Warning: line %d invalid format '%s', expected 'name,idType,url'\n", lineNum, line)
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
			fmt.Fprintf(os.Stderr, "Warning: line %d invalid idType '%s', using aggregateId\n", lineNum, idTypeStr)
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
