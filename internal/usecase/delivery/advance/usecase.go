package advance

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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
			return &Out{Completed: false}, nil
		}
		return nil, err
	}
	if o.Status != orderstatus.StatusInDelivery {
		return &Out{Completed: false}, nil
	}
	now := time.Now().UTC()
	if _, updateErr := uc.deliveries.UpdateStatus(ctx, deliveryrepo.UpdateStatusIn{
		OrderID: in.OrderID, Status: "delivered", DeliveredAt: &now,
	}); updateErr != nil {
		if pgerr.IsNotFound(updateErr) {
			return &Out{Completed: false}, nil
		}
		return nil, updateErr
	}
	payload, _ := json.Marshal(map[string]any{"order_id": in.OrderID.String()})
	_, err = uc.orders.ApplyTransition(ctx, orderrepo.TransitionIn{
		OrderID: in.OrderID, FromStatus: o.Status, ToStatus: orderstatus.StatusCompleted,
		ActorType: orderstatus.ActorSystem, ActorID: "delivery_simulator",
		FulfillmentType: o.FulfillmentType, OutboxEventType: "order.completed", OutboxPayload: payload,
	})
	if err != nil {
		if pgerr.IsNotFound(err) {
			return &Out{Completed: false}, nil
		}
		return nil, fmt.Errorf("complete order: %w", err)
	}
	return &Out{Completed: true}, nil
}
