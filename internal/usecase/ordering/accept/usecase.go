package accept

import (
	"context"
	"encoding/json"
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
	o, err := uc.orders.GetByID(ctx, in.OrderID)
	if err != nil {
		if pgerr.IsNotFound(err) {
			return nil, usecase.ErrOrderNotFound
		}
		return nil, err
	}
	if o.VenueID != in.VenueID {
		return nil, usecase.ErrOrderNotFound
	}
	tc := orderstatus.TransitionContext{
		Now: uc.nower.Now(), AcceptedAt: o.AcceptedAt, CancelGrace: uc.cancelGrace,
		AcceptTimeout: uc.acceptTimeout, PlacedAt: o.PlacedAt, FulfillmentType: o.FulfillmentType,
	}
	to := orderstatus.StatusAccepted
	if transErr := orderstatus.CanTransition(o.Status, to, orderstatus.ActorVenue, tc); transErr != nil {
		return nil, fmt.Errorf("%w: %v", usecase.ErrInvalidTransition, errdetails.InvalidTransitionDetail(o.Status, to))
	}
	var eta *time.Time
	if in.PrepMinutes != nil {
		t := uc.nower.Now().Add(time.Duration(*in.PrepMinutes) * time.Minute)
		eta = &t
	}
	payload, _ := json.Marshal(map[string]any{"order_id": in.OrderID.String(), "status": to})
	updated, err := uc.orders.ApplyTransition(ctx, orderrepo.TransitionIn{
		OrderID: in.OrderID, VenueID: &in.VenueID, FromStatus: o.Status, ToStatus: to,
		ActorType: orderstatus.ActorVenue, ActorID: in.VenueID.String(), FulfillmentType: o.FulfillmentType,
		OutboxEventType: "order.accepted", OutboxPayload: payload, PrepMinutes: in.PrepMinutes, ETAAt: eta,
	})
	if err != nil {
		if pgerr.IsNotFound(err) {
			return nil, usecase.ErrInvalidTransition
		}
		return nil, err
	}
	return &Out{Order: updated}, nil
}
