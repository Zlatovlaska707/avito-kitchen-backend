package apikey_issue

import (
	"context"

	apikeyissue "github.com/zlatovlaska707/avito-kitchen/internal/usecase/venues/apikey_issue"
)

type UseCase interface {
	Run(ctx context.Context, in apikeyissue.In) (*apikeyissue.Out, error)
}
