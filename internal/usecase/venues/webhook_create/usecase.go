package webhook_create

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/pgerr"
	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/webhook"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
	webhookrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/webhook"
)

type UseCase struct {
	webhooks webhookrepo.Repository
	ssrf     *webhook.SSRFGuard
}

func New(webhooks webhookrepo.Repository, ssrf *webhook.SSRFGuard) *UseCase {
	return &UseCase{webhooks: webhooks, ssrf: ssrf}
}

func (uc *UseCase) Run(ctx context.Context, in In) (*Out, error) {
	if err := uc.ssrf.ValidateURL(ctx, in.URL); err != nil {
		return nil, usecase.ErrWebhookURLForbidden
	}
	secret := in.Secret
	if secret == "" {
		b := make([]byte, 32)
		_, _ = rand.Read(b)
		secret = hex.EncodeToString(b)
	}
	wh, err := uc.webhooks.Create(ctx, webhookrepo.CreateIn{
		VenueID: in.VenueID, URL: in.URL, Secret: secret, EventTypes: in.EventTypes,
	})
	if err != nil {
		if pgerr.IsUniqueViolation(err) {
			return nil, usecase.ErrWebhookAlreadyExists
		}
		return nil, err
	}
	return &Out{Webhook: *wh}, nil
}
