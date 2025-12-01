package fetcher

import (
	"drill/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type FetchResult struct {
	Events   []models.Event
	Commands []models.Command
	Error    error
}

type Fetcher struct {
	client   *http.Client
	services []models.ServiceConfig
}

func NewFetcher(services []models.ServiceConfig) *Fetcher {
	return &Fetcher{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		services: services,
	}
}

func (f *Fetcher) FetchAll(aggregateID string) ([]models.Event, []models.Command, error) {
	var wg sync.WaitGroup
	resultsChan := make(chan FetchResult, len(f.services)*2)

	for _, service := range f.services {
		wg.Add(2)

		// Fetch events
		go func(svc models.ServiceConfig) {
			defer wg.Done()
			events, err := f.fetchEvents(svc, aggregateID)
			resultsChan <- FetchResult{Events: events, Error: err}
		}(service)

		// Fetch commands
		go func(svc models.ServiceConfig) {
			defer wg.Done()
			commands, err := f.fetchCommands(svc, aggregateID)
			resultsChan <- FetchResult{Commands: commands, Error: err}
		}(service)
	}

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	var allEvents []models.Event
	var allCommands []models.Command
	var errs []error

	for result := range resultsChan {
		if result.Error != nil {
			errs = append(errs, result.Error)
			continue
		}
		allEvents = append(allEvents, result.Events...)
		allCommands = append(allCommands, result.Commands...)
	}

	if len(errs) > 0 && len(allEvents) == 0 && len(allCommands) == 0 {
		return nil, nil, fmt.Errorf("all fetches failed: %v", errs)
	}

	return allEvents, allCommands, nil
}

func (f *Fetcher) fetchEvents(service models.ServiceConfig, aggregateID string) ([]models.Event, error) {
	url := fmt.Sprintf("%s/events?aggregateId=%s", service.URL, aggregateID)

	resp, err := f.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch events from %s: %w", service.Name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from %s", resp.StatusCode, service.Name)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response from %s: %w", service.Name, err)
	}

	var events []models.Event
	if err := json.Unmarshal(body, &events); err != nil {
		return nil, fmt.Errorf("failed to parse events from %s: %w", service.Name, err)
	}

	// Tag events with service name
	for i := range events {
		events[i].ServiceName = service.Name
	}

	return events, nil
}

func (f *Fetcher) fetchCommands(service models.ServiceConfig, aggregateID string) ([]models.Command, error) {
	url := fmt.Sprintf("%s/commandLifecycle?aggregateId=%s", service.URL, aggregateID)

	resp, err := f.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch commands from %s: %w", service.Name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from %s", resp.StatusCode, service.Name)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response from %s: %w", service.Name, err)
	}

	var commands []models.Command
	if err := json.Unmarshal(body, &commands); err != nil {
		return nil, fmt.Errorf("failed to parse commands from %s: %w", service.Name, err)
	}

	// Tag commands with service name
	for i := range commands {
		commands[i].ServiceName = service.Name
	}

	return commands, nil
}
