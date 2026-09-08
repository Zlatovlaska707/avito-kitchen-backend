package assign_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zlatovlaska707/avito-kitchen/internal/domain/orderstatus"
	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/pgerr"
	orderrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/order"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/delivery/assign"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/uctest"
)

func TestRun_HappyPathFromReady(t *testing.T) {
	orderID := uuid.New()
	acceptedAt := time.Now().Add(-10 * time.Minute)

	orders := &uctest.OrderRepo{
		Order: &orderrepo.Order{
			ID: orderID, Status: orderstatus.StatusReady,
			FulfillmentType: orderstatus.FulfillmentDelivery, PlacedAt: time.Now().Add(-20 * time.Minute),
			AcceptedAt: &acceptedAt,
		},
		TransitionOut: &orderrepo.Order{ID: orderID, Status: orderstatus.StatusInDelivery},
	}
	deliveries := &uctest.DeliveryRepo{}
	uc := assign.New(deliveries, orders)

	out, err := uc.Run(context.Background(), assign.In{
		OrderID: orderID, CourierName: "Ivan", CourierPhone: "+79001234567",
	})
	require.NoError(t, err)
	assert.True(t, out.Assigned)
}

func TestRun_HappyPathAlreadyInDelivery(t *testing.T) {
	orderID := uuid.New()
	orders := &uctest.OrderRepo{
		Order: &orderrepo.Order{
			ID: orderID, Status: orderstatus.StatusInDelivery,
			FulfillmentType: orderstatus.FulfillmentDelivery, PlacedAt: time.Now(),
		},
	}
	deliveries := &uctest.DeliveryRepo{}
	uc := assign.New(deliveries, orders)

	out, err := uc.Run(context.Background(), assign.In{
		OrderID: orderID, CourierName: "Petr", CourierPhone: "+79007654321",
	})
	require.NoError(t, err)
	assert.True(t, out.Assigned)
}

func TestRun_OrderNotFound(t *testing.T) {
	orders := &uctest.OrderRepo{GetByIDErr: pgerr.ErrNotFound}
	uc := assign.New(&uctest.DeliveryRepo{}, orders)

	out, err := uc.Run(context.Background(), assign.In{OrderID: uuid.New()})
	require.NoError(t, err)
	assert.False(t, out.Assigned)
}

func TestRun_WrongStatus(t *testing.T) {
	orderID := uuid.New()
	orders := &uctest.OrderRepo{
		Order: &orderrepo.Order{
			ID: orderID, Status: orderstatus.StatusPlaced,
			FulfillmentType: orderstatus.FulfillmentDelivery, PlacedAt: time.Now(),
		},
	}
	uc := assign.New(&uctest.DeliveryRepo{}, orders)

	out, err := uc.Run(context.Background(), assign.In{OrderID: orderID})
	require.NoError(t, err)
	assert.False(t, out.Assigned)
}

func TestRun_DeliveryNotFound(t *testing.T) {
	orderID := uuid.New()
	acceptedAt := time.Now().Add(-time.Minute)
	orders := &uctest.OrderRepo{
		Order: &orderrepo.Order{
			ID: orderID, Status: orderstatus.StatusReady,
			FulfillmentType: orderstatus.FulfillmentDelivery, PlacedAt: time.Now(),
			AcceptedAt: &acceptedAt,
		},
	}
	deliveries := &uctest.DeliveryRepo{UpdateErr: pgerr.ErrNotFound}
	uc := assign.New(deliveries, orders)

	out, err := uc.Run(context.Background(), assign.In{OrderID: orderID})
	require.NoError(t, err)
	assert.False(t, out.Assigned)
}
