package order_list

import (
	"context"

	orderlist "github.com/zlatovlaska707/avito-kitchen/internal/usecase/ordering/list_for_venue"
)

type UseCase interface {
	Run(ctx context.Context, in orderlist.In) (*orderlist.Out, error)
}
