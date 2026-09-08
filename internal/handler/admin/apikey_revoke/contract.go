package apikey_revoke

import (
	"context"

	apikeyrevoke "github.com/zlatovlaska707/avito-kitchen/internal/usecase/venues/apikey_revoke"
)

type UseCase interface {
	Run(ctx context.Context, in apikeyrevoke.In) (*apikeyrevoke.Out, error)
}
