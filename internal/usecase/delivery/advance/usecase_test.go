package advance_test

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
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/delivery/advance"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/uctest"
)

func TestRun_HappyPath(t *testing.T) {
	orderID := uuid.New()
	orders := &uctest.OrderRepo{
		Order: &orderrepo.Order{
			ID: orderID, Status: orderstatus.StatusInDelivery,
			FulfillmentType: orderstatus.FulfillmentDelivery, PlacedAt: time.Now().Add(-time.Hour),
		},
		TransitionOut: &orderrepo.Order{ID: orderID, Status: orderstatus.StatusCompleted},
	}
	deliveries := &uctest.DeliveryRepo{}
	uc := advance.New(deliveries, orders)

	out, err := uc.Run(context.Background(), advance.In{OrderID: orderID})
	require.NoError(t, err)
	assert.True(t, out.Completed)
}

func TestRun_OrderNotFound(t *testing.T) {
	orders := &uctest.OrderRepo{GetByIDErr: pgerr.ErrNotFound}
	uc := advance.New(&uctest.DeliveryRepo{}, orders)

	out, err := uc.Run(context.Background(), advance.In{OrderID: uuid.New()})
	require.NoError(t, err)
	assert.False(t, out.Completed)
}

func TestRun_WrongStatus(t *testing.T) {
	orderID := uuid.New()
	orders := &uctest.OrderRepo{
		Order: &orderrepo.Order{
			ID: orderID, Status: orderstatus.StatusReady,
			FulfillmentType: orderstatus.FulfillmentDelivery, PlacedAt: time.Now(),
		},
	}
	uc := advance.New(&uctest.DeliveryRepo{}, orders)

	out, err := uc.Run(context.Background(), advance.In{OrderID: orderID})
	require.NoError(t, err)
	assert.False(t, out.Completed)
}

func TestRun_DeliveryNotFound(t *testing.T) {
	orderID := uuid.New()
	orders := &uctest.OrderRepo{
		Order: &orderrepo.Order{
			ID: orderID, Status: orderstatus.StatusInDelivery,
			FulfillmentType: orderstatus.FulfillmentDelivery, PlacedAt: time.Now(),
		},
	}
	deliveries := &uctest.DeliveryRepo{UpdateErr: pgerr.ErrNotFound}
	uc := advance.New(deliveries, orders)

	out, err := uc.Run(context.Background(), advance.In{OrderID: orderID})
	require.NoError(t, err)
	assert.False(t, out.Completed)
}

func TestRun_TransitionRace(t *testing.T) {
	orderID := uuid.New()
	orders := &uctest.OrderRepo{
		Order: &orderrepo.Order{
			ID: orderID, Status: orderstatus.StatusInDelivery,
			FulfillmentType: orderstatus.FulfillmentDelivery, PlacedAt: time.Now(),
		},
		TransitionErr: pgerr.ErrNotFound,
	}
	deliveries := &uctest.DeliveryRepo{}
	uc := advance.New(deliveries, orders)

	out, err := uc.Run(context.Background(), advance.In{OrderID: orderID})
	require.NoError(t, err)
	assert.False(t, out.Completed)
}
