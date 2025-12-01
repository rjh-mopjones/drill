package cache

import (
	"drill/models"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const (
	MaxCachedRequests = 5
	CacheFileName     = ".drill_cache.json"
)

type CachedRequest struct {
	AggregateID string           `json:"aggregateId"`
	Timestamp   time.Time        `json:"timestamp"`
	Events      []models.Event   `json:"events"`
	Commands    []models.Command `json:"commands"`
	IsMock      bool             `json:"isMock"`
}

type Cache struct {
	Requests []CachedRequest `json:"requests"`
}

func getCachePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, CacheFileName), nil
}

func Load() (*Cache, error) {
	cachePath, err := getCachePath()
	if err != nil {
		return &Cache{}, nil
	}

	data, err := os.ReadFile(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Cache{}, nil
		}
		return nil, err
	}

	var cache Cache
	if err := json.Unmarshal(data, &cache); err != nil {
		return &Cache{}, nil
	}

	return &cache, nil
}

func (c *Cache) Save() error {
	cachePath, err := getCachePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cachePath, data, 0644)
}

func (c *Cache) AddRequest(aggregateID string, events []models.Event, commands []models.Command, isMock bool) {
	// Remove existing request with same aggregate ID
	filtered := make([]CachedRequest, 0)
	for _, r := range c.Requests {
		if r.AggregateID != aggregateID {
			filtered = append(filtered, r)
		}
	}
	c.Requests = filtered

	// Add new request at the beginning
	newRequest := CachedRequest{
		AggregateID: aggregateID,
		Timestamp:   time.Now(),
		Events:      events,
		Commands:    commands,
		IsMock:      isMock,
	}

	c.Requests = append([]CachedRequest{newRequest}, c.Requests...)

	// Keep only the last MaxCachedRequests
	if len(c.Requests) > MaxCachedRequests {
		c.Requests = c.Requests[:MaxCachedRequests]
	}
}

func (c *Cache) GetRequest(aggregateID string) *CachedRequest {
	for i := range c.Requests {
		if c.Requests[i].AggregateID == aggregateID {
			return &c.Requests[i]
		}
	}
	return nil
}

func (c *Cache) GetRecentRequests() []CachedRequest {
	// Sort by timestamp descending
	sorted := make([]CachedRequest, len(c.Requests))
	copy(sorted, c.Requests)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Timestamp.After(sorted[j].Timestamp)
	})
	return sorted
}
