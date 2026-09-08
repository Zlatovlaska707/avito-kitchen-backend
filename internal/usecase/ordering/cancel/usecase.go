package cancel

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/zlatovlaska707/avito-kitchen/internal/domain/orderstatus"
	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/pgerr"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
	nowercontract "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/nower"
	orderrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/order"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/errdetails"
)

type UseCase struct {
	orders        orderrepo.Repository
	nower         nowercontract.Nower
	cancelGrace   time.Duration
	acceptTimeout time.Duration
}

func New(orders orderrepo.Repository, nower nowercontract.Nower, cancelGrace, acceptTimeout time.Duration) *UseCase {
	return &UseCase{orders: orders, nower: nower, cancelGrace: cancelGrace, acceptTimeout: acceptTimeout}
}

func (uc *UseCase) Run(ctx context.Context, in In) (*Out, error) {
	o, err := uc.orders.GetByIDForCustomer(ctx, in.OrderID, in.CustomerID)
	if err != nil {
		if pgerr.IsNotFound(err) {
			return nil, usecase.ErrOrderNotFound
		}
		return nil, err
	}
	tc := orderstatus.TransitionContext{
		Now: uc.nower.Now(), AcceptedAt: o.AcceptedAt, CancelGrace: uc.cancelGrace,
		AcceptTimeout: uc.acceptTimeout, PlacedAt: o.PlacedAt, FulfillmentType: o.FulfillmentType,
	}
	to := orderstatus.StatusCancelled
	if transErr := orderstatus.CanTransition(o.Status, to, in.ActorType, tc); transErr != nil {
		if errors.Is(transErr, orderstatus.ErrNotCancellable) {
			return nil, usecase.ErrNotCancellable
		}
		return nil, fmt.Errorf("%w: %v", usecase.ErrInvalidTransition, errdetails.InvalidTransitionDetail(o.Status, to))
	}
	cancelledBy := in.ActorID
	payload, _ := json.Marshal(map[string]any{"order_id": in.OrderID.String(), "status": to})
	updated, err := uc.orders.ApplyTransition(ctx, orderrepo.TransitionIn{
		OrderID: in.OrderID, FromStatus: o.Status, ToStatus: to,
		ActorType: in.ActorType, ActorID: in.ActorID, Reason: in.Reason,
		CancelledBy: &cancelledBy, FulfillmentType: o.FulfillmentType,
		OutboxEventType: "order.cancelled", OutboxPayload: payload,
	})
	if err != nil {
		if pgerr.IsNotFound(err) {
			return nil, usecase.ErrInvalidTransition
		}
		return nil, err
	}
	return &Out{Order: updated}, nil
}
