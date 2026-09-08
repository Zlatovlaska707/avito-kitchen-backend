package webhook_create_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/webhook"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
	venuewebhookcreate "github.com/zlatovlaska707/avito-kitchen/internal/usecase/venues/webhook_create"
	webhookrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/webhook"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/uctest"
)

func ssrfGuard() *webhook.SSRFGuard {
	return &webhook.SSRFGuard{AllowPrivateIPs: true}
}

func TestRun_ForbiddenURL(t *testing.T) {
	uc := venuewebhookcreate.New(&uctest.WebhookRepo{}, ssrfGuard())

	_, err := uc.Run(context.Background(), venuewebhookcreate.In{
		VenueID: uuid.New(), URL: "http://localhost:8080/hook", EventTypes: []string{"order.placed"},
	})
	assert.ErrorIs(t, err, usecase.ErrWebhookURLForbidden)
}

func TestRun_HappyPath(t *testing.T) {
	venueID := uuid.New()
	webhookID := uuid.New()
	repo := &uctest.WebhookRepo{Webhook: &webhookrepo.Webhook{
		ID: webhookID, VenueID: venueID, URL: "http://192.168.1.10:8080/hook",
		Secret: "provided-secret", EventTypes: []string{"order.placed"}, IsActive: true,
	}}
	uc := venuewebhookcreate.New(repo, ssrfGuard())

	out, err := uc.Run(context.Background(), venuewebhookcreate.In{
		VenueID: venueID, URL: "http://192.168.1.10:8080/hook",
		Secret: "provided-secret", EventTypes: []string{"order.placed"},
	})
	require.NoError(t, err)
	assert.Equal(t, webhookID, out.Webhook.ID)
	assert.Equal(t, "provided-secret", out.Webhook.Secret)
}

func TestRun_AutoGeneratesSecret(t *testing.T) {
	venueID := uuid.New()
	repo := &uctest.WebhookRepo{Webhook: &webhookrepo.Webhook{
		ID: uuid.New(), VenueID: venueID, URL: "http://10.0.0.5:443/hook",
		Secret: "generated", EventTypes: []string{"order.ready"},
	}}
	uc := venuewebhookcreate.New(repo, ssrfGuard())

	out, err := uc.Run(context.Background(), venuewebhookcreate.In{
		VenueID: venueID, URL: "http://10.0.0.5:443/hook", EventTypes: []string{"order.ready"},
	})
	require.NoError(t, err)
	assert.NotEmpty(t, out.Webhook.Secret)
}

func TestRun_AlreadyExists(t *testing.T) {
	uc := venuewebhookcreate.New(
		&uctest.WebhookRepo{CreateErr: &pgconn.PgError{Code: "23505"}},
		ssrfGuard(),
	)

	_, err := uc.Run(context.Background(), venuewebhookcreate.In{
		VenueID: uuid.New(), URL: "http://192.168.0.1:8080/hook", EventTypes: []string{"order.placed"},
	})
	assert.ErrorIs(t, err, usecase.ErrWebhookAlreadyExists)
}
