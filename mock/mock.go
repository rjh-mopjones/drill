package mock

import (
	"drill/models"
	"time"

	"github.com/google/uuid"
)

var MockServices = []models.ServiceConfig{
	{Name: "account-service", IDType: models.IDTypeAggregate, URL: "http://localhost:8081"},
	{Name: "payment-service", IDType: models.IDTypeIndex, URL: "http://localhost:8082"},
	{Name: "notification-service", IDType: models.IDTypeAggregate, URL: "http://localhost:8083"},
	{Name: "audit-service", IDType: models.IDTypeIndex, URL: "http://localhost:8084"},
	{Name: "billing-service", IDType: models.IDTypeAggregate, URL: "http://localhost:8085"},
}

func GenerateMockData(aggregateID string) ([]models.Event, []models.Command) {
	baseTime := time.Now().Add(-24 * time.Hour)

	// Generate some shared correlation IDs for linking commands and events
	correlationIDs := []string{
		uuid.New().String(),
		uuid.New().String(),
		uuid.New().String(),
		uuid.New().String(),
	}

	events := []models.Event{
		// Account service events
		{
			Metadata: models.EventMetadata{
				EventID:       uuid.New().String(),
				EventAlias:    "AccountCreated",
				PersistedAt:   baseTime.Add(1 * time.Minute),
				CorrelationID: correlationIDs[0],
				AggregateID:   aggregateID,
			},
			Payload:     `{"name": "John Doe", "email": "john@example.com"}`,
			ServiceName: "account-service",
		},
		{
			Metadata: models.EventMetadata{
				EventID:       uuid.New().String(),
				EventAlias:    "AccountVerified",
				PersistedAt:   baseTime.Add(5 * time.Minute),
				CorrelationID: correlationIDs[0],
				AggregateID:   aggregateID,
			},
			Payload:     `{"verifiedAt": "2024-01-15T10:00:00Z"}`,
			ServiceName: "account-service",
		},
		{
			Metadata: models.EventMetadata{
				EventID:       uuid.New().String(),
				EventAlias:    "ProfileUpdated",
				PersistedAt:   baseTime.Add(30 * time.Minute),
				CorrelationID: correlationIDs[1],
				AggregateID:   aggregateID,
			},
			Payload:     `{"field": "address", "value": "123 Main St"}`,
			ServiceName: "account-service",
		},
		// Payment service events
		{
			Metadata: models.EventMetadata{
				EventID:       uuid.New().String(),
				EventAlias:    "PaymentMethodAdded",
				PersistedAt:   baseTime.Add(10 * time.Minute),
				CorrelationID: correlationIDs[1],
				AggregateID:   aggregateID,
			},
			Payload:     `{"type": "credit_card", "last4": "4242"}`,
			ServiceName: "payment-service",
		},
		{
			Metadata: models.EventMetadata{
				EventID:       uuid.New().String(),
				EventAlias:    "PaymentProcessed",
				PersistedAt:   baseTime.Add(45 * time.Minute),
				CorrelationID: correlationIDs[2],
				AggregateID:   aggregateID,
			},
			Payload:     `{"amount": 99.99, "currency": "USD"}`,
			ServiceName: "payment-service",
		},
		{
			Metadata: models.EventMetadata{
				EventID:       uuid.New().String(),
				EventAlias:    "RefundIssued",
				PersistedAt:   baseTime.Add(2 * time.Hour),
				CorrelationID: correlationIDs[3],
				AggregateID:   aggregateID,
			},
			Payload:     `{"amount": 25.00, "reason": "partial_refund"}`,
			ServiceName: "payment-service",
		},
		// Notification service events
		{
			Metadata: models.EventMetadata{
				EventID:       uuid.New().String(),
				EventAlias:    "WelcomeEmailSent",
				PersistedAt:   baseTime.Add(2 * time.Minute),
				CorrelationID: correlationIDs[0],
				AggregateID:   aggregateID,
			},
			Payload:     `{"template": "welcome", "recipient": "john@example.com"}`,
			ServiceName: "notification-service",
		},
		{
			Metadata: models.EventMetadata{
				EventID:       uuid.New().String(),
				EventAlias:    "PaymentReceiptSent",
				PersistedAt:   baseTime.Add(46 * time.Minute),
				CorrelationID: correlationIDs[2],
				AggregateID:   aggregateID,
			},
			Payload:     `{"template": "receipt", "amount": 99.99}`,
			ServiceName: "notification-service",
		},
		// Audit service events
		{
			Metadata: models.EventMetadata{
				EventID:       uuid.New().String(),
				EventAlias:    "AuditLogCreated",
				PersistedAt:   baseTime.Add(1*time.Minute + 30*time.Second),
				CorrelationID: correlationIDs[0],
				AggregateID:   aggregateID,
			},
			Payload:     `{"action": "account.create", "actor": "system"}`,
			ServiceName: "audit-service",
		},
		{
			Metadata: models.EventMetadata{
				EventID:       uuid.New().String(),
				EventAlias:    "ComplianceCheckPassed",
				PersistedAt:   baseTime.Add(3 * time.Minute),
				CorrelationID: correlationIDs[0],
				AggregateID:   aggregateID,
			},
			Payload:     `{"checkType": "kyc", "status": "passed"}`,
			ServiceName: "audit-service",
		},
		// Billing service events
		{
			Metadata: models.EventMetadata{
				EventID:       uuid.New().String(),
				EventAlias:    "SubscriptionCreated",
				PersistedAt:   baseTime.Add(15 * time.Minute),
				CorrelationID: correlationIDs[1],
				AggregateID:   aggregateID,
			},
			Payload:     `{"plan": "premium", "interval": "monthly"}`,
			ServiceName: "billing-service",
		},
		{
			Metadata: models.EventMetadata{
				EventID:       uuid.New().String(),
				EventAlias:    "InvoiceGenerated",
				PersistedAt:   baseTime.Add(44 * time.Minute),
				CorrelationID: correlationIDs[2],
				AggregateID:   aggregateID,
			},
			Payload:     `{"invoiceId": "INV-001", "amount": 99.99}`,
			ServiceName: "billing-service",
		},
	}

	commands := []models.Command{
		// Account service commands
		{
			CommandID:     uuid.New().String(),
			CommandStatus: models.ExecutionSucceeded,
			CommandAlias:  "CreateAccount",
			PersistedAt:   baseTime.Add(30 * time.Second),
			Payload:       `{"name": "John Doe", "email": "john@example.com"}`,
			CorrelationID: correlationIDs[0],
			AggregateID:   aggregateID,
			ServiceName:   "account-service",
		},
		{
			CommandID:     uuid.New().String(),
			CommandStatus: models.ExecutionSucceeded,
			CommandAlias:  "VerifyAccount",
			PersistedAt:   baseTime.Add(4 * time.Minute),
			Payload:       `{"verificationCode": "123456"}`,
			CorrelationID: correlationIDs[0],
			AggregateID:   aggregateID,
			ServiceName:   "account-service",
		},
		{
			CommandID:     uuid.New().String(),
			CommandStatus: models.ExecutionSucceeded,
			CommandAlias:  "UpdateProfile",
			PersistedAt:   baseTime.Add(29 * time.Minute),
			Payload:       `{"address": "123 Main St"}`,
			CorrelationID: correlationIDs[1],
			AggregateID:   aggregateID,
			ServiceName:   "account-service",
		},
		// Payment service commands
		{
			CommandID:     uuid.New().String(),
			CommandStatus: models.ExecutionSucceeded,
			CommandAlias:  "AddPaymentMethod",
			PersistedAt:   baseTime.Add(9 * time.Minute),
			Payload:       `{"type": "credit_card", "token": "tok_xxx"}`,
			CorrelationID: correlationIDs[1],
			AggregateID:   aggregateID,
			ServiceName:   "payment-service",
		},
		{
			CommandID:     uuid.New().String(),
			CommandStatus: models.ExecutionSucceeded,
			CommandAlias:  "ProcessPayment",
			PersistedAt:   baseTime.Add(44 * time.Minute),
			Payload:       `{"amount": 99.99, "currency": "USD"}`,
			CorrelationID: correlationIDs[2],
			AggregateID:   aggregateID,
			ServiceName:   "payment-service",
		},
		{
			CommandID:     uuid.New().String(),
			CommandStatus: models.CommandFailed,
			CommandAlias:  "ProcessRefund",
			PersistedAt:   baseTime.Add(1*time.Hour + 50*time.Minute),
			Payload:       `{"amount": 50.00, "reason": "customer_request"}`,
			CorrelationID: correlationIDs[3],
			AggregateID:   aggregateID,
			ServiceName:   "payment-service",
		},
		{
			CommandID:     uuid.New().String(),
			CommandStatus: models.ExecutionSucceeded,
			CommandAlias:  "ProcessRefund",
			PersistedAt:   baseTime.Add(1*time.Hour + 55*time.Minute),
			Payload:       `{"amount": 25.00, "reason": "partial_refund"}`,
			CorrelationID: correlationIDs[3],
			AggregateID:   aggregateID,
			ServiceName:   "payment-service",
		},
		// Notification service commands
		{
			CommandID:     uuid.New().String(),
			CommandStatus: models.ExecutionSucceeded,
			CommandAlias:  "SendEmail",
			PersistedAt:   baseTime.Add(1*time.Minute + 45*time.Second),
			Payload:       `{"template": "welcome", "to": "john@example.com"}`,
			CorrelationID: correlationIDs[0],
			AggregateID:   aggregateID,
			ServiceName:   "notification-service",
		},
		{
			CommandID:     uuid.New().String(),
			CommandStatus: models.CommandFailed,
			CommandAlias:  "SendSMS",
			PersistedAt:   baseTime.Add(6 * time.Minute),
			Payload:       `{"to": "+1234567890", "message": "Welcome!"}`,
			CorrelationID: correlationIDs[0],
			AggregateID:   aggregateID,
			ServiceName:   "notification-service",
		},
		// Audit service commands
		{
			CommandID:     uuid.New().String(),
			CommandStatus: models.ExecutionSucceeded,
			CommandAlias:  "CreateAuditLog",
			PersistedAt:   baseTime.Add(1*time.Minute + 15*time.Second),
			Payload:       `{"action": "account.create"}`,
			CorrelationID: correlationIDs[0],
			AggregateID:   aggregateID,
			ServiceName:   "audit-service",
		},
		{
			CommandID:     uuid.New().String(),
			CommandStatus: models.ExecutionSucceeded,
			CommandAlias:  "RunComplianceCheck",
			PersistedAt:   baseTime.Add(2*time.Minute + 30*time.Second),
			Payload:       `{"checkType": "kyc"}`,
			CorrelationID: correlationIDs[0],
			AggregateID:   aggregateID,
			ServiceName:   "audit-service",
		},
		// Billing service commands
		{
			CommandID:     uuid.New().String(),
			CommandStatus: models.ExecutionSucceeded,
			CommandAlias:  "CreateSubscription",
			PersistedAt:   baseTime.Add(14 * time.Minute),
			Payload:       `{"plan": "premium"}`,
			CorrelationID: correlationIDs[1],
			AggregateID:   aggregateID,
			ServiceName:   "billing-service",
		},
		{
			CommandID:     uuid.New().String(),
			CommandStatus: models.ExecutionSucceeded,
			CommandAlias:  "GenerateInvoice",
			PersistedAt:   baseTime.Add(43 * time.Minute),
			Payload:       `{"subscriptionId": "sub_xxx"}`,
			CorrelationID: correlationIDs[2],
			AggregateID:   aggregateID,
			ServiceName:   "billing-service",
		},
		{
			CommandID:     uuid.New().String(),
			CommandStatus: models.CommandFailed,
			CommandAlias:  "CancelSubscription",
			PersistedAt:   baseTime.Add(3 * time.Hour),
			Payload:       `{"reason": "customer_request"}`,
			CorrelationID: correlationIDs[3],
			AggregateID:   aggregateID,
			ServiceName:   "billing-service",
		},
	}

	return events, commands
}
