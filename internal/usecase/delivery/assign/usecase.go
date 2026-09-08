package assign

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/zlatovlaska707/avito-kitchen/internal/domain/orderstatus"
	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/pgerr"
	deliveryrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/delivery"
	orderrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/order"
)

type UseCase struct {
	deliveries deliveryrepo.Repository
	orders     orderrepo.Repository
}

func New(deliveries deliveryrepo.Repository, orders orderrepo.Repository) *UseCase {
	return &UseCase{deliveries: deliveries, orders: orders}
}

func (uc *UseCase) Run(ctx context.Context, in In) (*Out, error) {
	o, err := uc.orders.GetByID(ctx, in.OrderID)
	if err != nil {
		if pgerr.IsNotFound(err) {
			return &Out{Assigned: false}, nil
		}
		return nil, err
	}
	if o.Status != orderstatus.StatusReady && o.Status != orderstatus.StatusInDelivery {
		return &Out{Assigned: false}, nil
	}
	name, phone := in.CourierName, in.CourierPhone
	if _, updateErr := uc.deliveries.UpdateStatus(ctx, deliveryrepo.UpdateStatusIn{
		OrderID: in.OrderID, Status: "assigned", CourierName: &name, CourierPhone: &phone,
	}); updateErr != nil {
		if pgerr.IsNotFound(updateErr) {
			return &Out{Assigned: false}, nil
		}
		return nil, updateErr
	}
	if o.Status == orderstatus.StatusReady {
		payload, _ := json.Marshal(map[string]any{"order_id": in.OrderID.String()})
		_, err = uc.orders.ApplyTransition(ctx, orderrepo.TransitionIn{
			OrderID: in.OrderID, FromStatus: o.Status, ToStatus: orderstatus.StatusInDelivery,
			ActorType: orderstatus.ActorSystem, ActorID: "delivery_simulator",
			FulfillmentType: o.FulfillmentType, OutboxEventType: "order.delivery_assigned", OutboxPayload: payload,
		})
		if err != nil && !pgerr.IsNotFound(err) {
			return nil, fmt.Errorf("transition order in_delivery: %w", err)
		}
	}
	return &Out{Assigned: true}, nil
}
