package venues_list

import (
	"context"

	venueslist "github.com/zlatovlaska707/avito-kitchen/internal/usecase/venues/list"
)

type UseCase interface {
	Run(ctx context.Context, in venueslist.In) (*venueslist.Out, error)
}
