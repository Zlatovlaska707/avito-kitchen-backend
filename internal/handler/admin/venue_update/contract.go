package venue_update

import (
	"context"

	venueupdate "github.com/zlatovlaska707/avito-kitchen/internal/usecase/venues/update"
)

type UseCase interface {
	Run(ctx context.Context, in venueupdate.In) (*venueupdate.Out, error)
}
