package models

import (
	"time"
)

type CommandStatus string

const (
	CommandFailed      CommandStatus = "COMMAND_FAILED"
	ExecutionSucceeded CommandStatus = "EXECUTION_SUCCEEDED"
)

type EventMetadata struct {
	EventID       string    `json:"eventId"`
	EventAlias    string    `json:"eventAlias"`
	PersistedAt   time.Time `json:"persistedAt"`
	CorrelationID string    `json:"correlationId"`
	AggregateID   string    `json:"aggregateId"`
}

type Event struct {
	Metadata    EventMetadata `json:"metadata"`
	Payload     string        `json:"payload"`
	ServiceName string        `json:"-"` // Added to track which service this came from
}

type Command struct {
	CommandID     string        `json:"commandId"`
	CommandStatus CommandStatus `json:"commandStatus"`
	CommandAlias  string        `json:"commandAlias"`
	PersistedAt   time.Time     `json:"persistedAt"`
	Payload       string        `json:"payload"`
	CorrelationID string        `json:"correlationId"`
	AggregateID   string        `json:"aggregateId"`
	ServiceName   string        `json:"-"` // Added to track which service this came from
}

type IDType string

const (
	IDTypeAggregate IDType = "aggregateId"
	IDTypeIndex     IDType = "indexId"
)

type ServiceConfig struct {
	Name   string
	IDType IDType
	URL    string
}
