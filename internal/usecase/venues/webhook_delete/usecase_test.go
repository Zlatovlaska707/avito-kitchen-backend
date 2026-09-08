package webhook_delete_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/pgerr"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/uctest"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/venues/webhook_delete"
)

func TestRun_HappyPath(t *testing.T) {
	venueID := uuid.New()
	webhookID := uuid.New()
	webhooks := &uctest.WebhookRepo{}
	uc := webhook_delete.New(webhooks)

	out, err := uc.Run(context.Background(), webhook_delete.In{
		VenueID:   venueID,
		WebhookID: webhookID,
	})
	require.NoError(t, err)
	require.NotNil(t, out)
}

func TestRun_WebhookNotFound(t *testing.T) {
	webhooks := &uctest.WebhookRepo{DeleteErr: pgerr.ErrNotFound}
	uc := webhook_delete.New(webhooks)

	_, err := uc.Run(context.Background(), webhook_delete.In{
		VenueID:   uuid.New(),
		WebhookID: uuid.New(),
	})
	assert.ErrorIs(t, err, usecase.ErrVenueNotFound)
}

func TestRun_DeleteError(t *testing.T) {
	webhooks := &uctest.WebhookRepo{DeleteErr: assert.AnError}
	uc := webhook_delete.New(webhooks)

	_, err := uc.Run(context.Background(), webhook_delete.In{
		VenueID:   uuid.New(),
		WebhookID: uuid.New(),
	})
	assert.ErrorIs(t, err, assert.AnError)
}
