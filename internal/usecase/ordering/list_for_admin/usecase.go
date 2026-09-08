package list_for_admin

import (
	"context"

	"github.com/zlatovlaska707/avito-kitchen/internal/domain/orderstatus"
	orderrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/order"
)

type UseCase struct {
	orders orderrepo.Repository
}

func New(orders orderrepo.Repository) *UseCase {
	return &UseCase{orders: orders}
}

func (uc *UseCase) Run(ctx context.Context, in In) (*Out, error) {
	limit := in.Limit
	if limit <= 0 {
		limit = 20
	}
	filter := orderrepo.ListFilter{Limit: limit + 1, CursorTS: in.CursorTS, CursorID: in.CursorID, Offset: in.Offset}
	if in.VenueID != nil {
		filter.VenueID = in.VenueID
	}
	if in.Status != nil {
		st := orderstatus.Status(*in.Status)
		filter.Status = &st
	}
	orders, err := uc.orders.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	hasNext := len(orders) > limit
	if hasNext {
		orders = orders[:limit]
	}
	out := &Out{Orders: orders, HasNext: hasNext}
	if len(orders) > 0 {
		last := orders[len(orders)-1]
		ts := last.CreatedAt
		id := last.ID
		out.LastTS = &ts
		out.LastID = &id
	}
	return out, nil
}
