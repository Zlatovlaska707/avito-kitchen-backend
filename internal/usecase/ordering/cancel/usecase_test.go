package cancel_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zlatovlaska707/avito-kitchen/internal/domain/orderstatus"
	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/pgerr"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
	orderrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/order"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/ordering/cancel"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/uctest"
)

func TestRun_HappyPath(t *testing.T) {
	now := time.Date(2026, time.September, 8, 12, 0, 0, 0, time.UTC)
	orderID := uuid.New()
	customerID := uuid.New()

	orders := &uctest.OrderRepo{
		Order: &orderrepo.Order{
			ID: orderID, CustomerID: customerID, Status: orderstatus.StatusPlaced,
			FulfillmentType: orderstatus.FulfillmentPickup, PlacedAt: now,
		},
		TransitionOut: &orderrepo.Order{ID: orderID, Status: orderstatus.StatusCancelled},
	}
	uc := cancel.New(orders, uctest.FixedNower{T: now}, 5*time.Minute, 10*time.Minute)

	out, err := uc.Run(context.Background(), cancel.In{
		OrderID: orderID, CustomerID: customerID,
		ActorType: orderstatus.ActorCustomer, ActorID: customerID.String(),
	})
	require.NoError(t, err)
	require.NotNil(t, out.Order)
	assert.Equal(t, orderstatus.StatusCancelled, out.Order.Status)
}

func TestRun_OrderNotFound(t *testing.T) {
	orders := &uctest.OrderRepo{GetByCustomerErr: pgerr.ErrNotFound}
	uc := cancel.New(orders, uctest.FixedNower{T: time.Now()}, time.Minute, 10*time.Minute)

	_, err := uc.Run(context.Background(), cancel.In{
		OrderID: uuid.New(), CustomerID: uuid.New(),
		ActorType: orderstatus.ActorCustomer, ActorID: "customer",
	})
	assert.ErrorIs(t, err, usecase.ErrOrderNotFound)
}

func TestRun_NotCancellable(t *testing.T) {
	now := time.Date(2026, time.September, 8, 12, 0, 0, 0, time.UTC)
	orderID := uuid.New()
	customerID := uuid.New()
	acceptedAt := now.Add(-10 * time.Minute)

	orders := &uctest.OrderRepo{
		Order: &orderrepo.Order{
			ID: orderID, CustomerID: customerID, Status: orderstatus.StatusAccepted,
			FulfillmentType: orderstatus.FulfillmentDelivery, PlacedAt: now.Add(-15 * time.Minute),
			AcceptedAt: &acceptedAt,
		},
	}
	uc := cancel.New(orders, uctest.FixedNower{T: now}, 5*time.Minute, 10*time.Minute)

	_, err := uc.Run(context.Background(), cancel.In{
		OrderID: orderID, CustomerID: customerID,
		ActorType: orderstatus.ActorCustomer, ActorID: customerID.String(),
	})
	assert.ErrorIs(t, err, usecase.ErrNotCancellable)
}

func TestRun_InvalidTransition(t *testing.T) {
	orderID := uuid.New()
	customerID := uuid.New()
	orders := &uctest.OrderRepo{
		Order: &orderrepo.Order{
			ID: orderID, CustomerID: customerID, Status: orderstatus.StatusCompleted,
			FulfillmentType: orderstatus.FulfillmentDelivery, PlacedAt: time.Now(),
		},
	}
	uc := cancel.New(orders, uctest.FixedNower{T: time.Now()}, time.Minute, 10*time.Minute)

	_, err := uc.Run(context.Background(), cancel.In{
		OrderID: orderID, CustomerID: customerID,
		ActorType: orderstatus.ActorCustomer, ActorID: customerID.String(),
	})
	assert.ErrorIs(t, err, usecase.ErrInvalidTransition)
}

func TestRun_TransitionRace(t *testing.T) {
	orderID := uuid.New()
	customerID := uuid.New()
	orders := &uctest.OrderRepo{
		Order: &orderrepo.Order{
			ID: orderID, CustomerID: customerID, Status: orderstatus.StatusPlaced,
			FulfillmentType: orderstatus.FulfillmentPickup, PlacedAt: time.Now(),
		},
		TransitionErr: pgerr.ErrNotFound,
	}
	uc := cancel.New(orders, uctest.FixedNower{T: time.Now()}, time.Minute, 10*time.Minute)

	_, err := uc.Run(context.Background(), cancel.In{
		OrderID: orderID, CustomerID: customerID,
		ActorType: orderstatus.ActorCustomer, ActorID: customerID.String(),
	})
	assert.ErrorIs(t, err, usecase.ErrInvalidTransition)
}
