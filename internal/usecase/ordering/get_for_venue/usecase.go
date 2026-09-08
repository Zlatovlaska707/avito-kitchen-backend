package get_for_venue

import (
	"context"

	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/pgerr"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
	deliveryrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/delivery"
	orderrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/order"
)

type UseCase struct {
	orders     orderrepo.Repository
	deliveries deliveryrepo.Repository
}

func New(orders orderrepo.Repository, deliveries deliveryrepo.Repository) *UseCase {
	return &UseCase{orders: orders, deliveries: deliveries}
}

func (uc *UseCase) Run(ctx context.Context, in In) (*Out, error) {
	o, err := uc.orders.GetByIDForVenue(ctx, in.OrderID, in.VenueID)
	if err != nil {
		if pgerr.IsNotFound(err) {
			return nil, usecase.ErrOrderNotFound
		}
		return nil, err
	}
	items, err := uc.orders.GetItems(ctx, o.ID)
	if err != nil {
		return nil, err
	}
	history, err := uc.orders.GetStatusHistory(ctx, o.ID)
	if err != nil {
		return nil, err
	}
	del, _ := uc.deliveries.GetByOrderID(ctx, o.ID)
	return &Out{Order: o, Items: items, History: history, Delivery: del}, nil
}
