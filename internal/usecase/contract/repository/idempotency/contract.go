package idempotency

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	Claim(ctx context.Context, in ClaimIn) (ClaimResult, error)
	Get(ctx context.Context, scope, key string) (*Record, error)
	SaveResponse(ctx context.Context, in SaveResponseIn) error
	Delete(ctx context.Context, scope, key string) error
	DeleteExpired(ctx context.Context) (int, error)
}

type ClaimIn struct {
	Scope       string
	Key         string
	RequestHash string
	ExpiresAt   time.Time
}

type ClaimResult int

const (
	ClaimGranted ClaimResult = iota
	ClaimExists
)

type Record struct {
	Scope          string
	Key            string
	RequestHash    string
	ResponseStatus *int
	ResponseBody   json.RawMessage
	ResourceID     *uuid.UUID
	CreatedAt      time.Time
	ExpiresAt      time.Time
}

type SaveResponseIn struct {
	Scope          string
	Key            string
	ResponseStatus int
	ResponseBody   json.RawMessage
	ResourceID     *uuid.UUID
}
