package outbox

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	Publish(ctx context.Context, in PublishIn) (int64, error)
	FetchPendingWithLease(ctx context.Context, batchSize int, lease time.Duration) ([]Event, error)
	MarkPublished(ctx context.Context, id int64) error
	MarkFailed(ctx context.Context, id int64, attempts int, nextAttemptAt time.Time, lastError string) error
	MarkSkipped(ctx context.Context, id int64) error
	RecordDelivery(ctx context.Context, in DeliveryIn) error
}

type Event struct {
	ID            int64
	AggregateType string
	AggregateID   uuid.UUID
	VenueID       *uuid.UUID
	EventType     string
	Payload       json.RawMessage
	Status        string
	Attempts      int
	NextAttemptAt time.Time
	CreatedAt     time.Time
}

type PublishIn struct {
	AggregateType string
	AggregateID   uuid.UUID
	VenueID       *uuid.UUID
	EventType     string
	Payload       json.RawMessage
}

type DeliveryIn struct {
	OutboxEventID int64
	WebhookID     *uuid.UUID
	Attempt       int
	HTTPStatus    *int
	DurationMS    *int
	Error         *string
}
