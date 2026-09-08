package expire

import (
	"context"
	"encoding/json"
	"time"

	"github.com/zlatovlaska707/avito-kitchen/internal/domain/orderstatus"
	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/pgerr"
	nowercontract "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/nower"
	orderrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/order"
)

type UseCase struct {
	orders        orderrepo.Repository
	nower         nowercontract.Nower
	acceptTimeout time.Duration
}

func New(orders orderrepo.Repository, nower nowercontract.Nower, acceptTimeout time.Duration) *UseCase {
	return &UseCase{orders: orders, nower: nower, acceptTimeout: acceptTimeout}
}

func (uc *UseCase) Run(ctx context.Context, in In) (*Out, error) {
	o, err := uc.orders.GetByID(ctx, in.OrderID)
	if err != nil {
		if pgerr.IsNotFound(err) {
			return &Out{Processed: false}, nil
		}
		return nil, err
	}
	if o.Status != orderstatus.StatusPlaced {
		return &Out{Processed: false}, nil
	}
	reason := "venue_no_response"
	tc := orderstatus.TransitionContext{
		Now: uc.nower.Now(), PlacedAt: o.PlacedAt, AcceptTimeout: uc.acceptTimeout,
		SystemTimeoutHit: true, FulfillmentType: o.FulfillmentType, Reason: reason,
	}
	to := orderstatus.StatusCancelled
	allowed := orderstatus.CanTransition(o.Status, to, orderstatus.ActorSystem, tc) == nil
	if !allowed {
		return &Out{Processed: false}, nil
	}
	cancelledBy := "system"
	payload, _ := json.Marshal(map[string]any{"order_id": in.OrderID.String(), "status": to})
	_, err = uc.orders.ApplyTransition(ctx, orderrepo.TransitionIn{
		OrderID: in.OrderID, FromStatus: o.Status, ToStatus: to,
		ActorType: orderstatus.ActorSystem, ActorID: "system", Reason: &reason,
		CancelledBy: &cancelledBy, FulfillmentType: o.FulfillmentType,
		OutboxEventType: "order.cancelled", OutboxPayload: payload,
	})
	if err != nil {
		if pgerr.IsNotFound(err) {
			return &Out{Processed: false}, nil
		}
		return nil, err
	}
	return &Out{Processed: true}, nil
}
