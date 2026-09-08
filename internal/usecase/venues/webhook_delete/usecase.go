package webhook_delete

import (
	"context"

	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/pgerr"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
	webhookrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/webhook"
)

type UseCase struct {
	webhooks webhookrepo.Repository
}

func New(webhooks webhookrepo.Repository) *UseCase {
	return &UseCase{webhooks: webhooks}
}

func (uc *UseCase) Run(ctx context.Context, in In) (*Out, error) {
	if err := uc.webhooks.Delete(ctx, in.VenueID, in.WebhookID); err != nil {
		if pgerr.IsNotFound(err) {
			return nil, usecase.ErrVenueNotFound
		}
		return nil, err
	}
	return &Out{}, nil
}
