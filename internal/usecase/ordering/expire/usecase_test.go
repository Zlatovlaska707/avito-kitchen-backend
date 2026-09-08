package expire_test

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
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/ordering/expire"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/uctest"
)

func TestRun_HappyPath(t *testing.T) {
	now := time.Date(2026, time.September, 8, 12, 0, 0, 0, time.UTC)
	orderID := uuid.New()

	orders := &uctest.OrderRepo{
		Order: &orderrepo.Order{
			ID: orderID, Status: orderstatus.StatusPlaced,
			FulfillmentType: orderstatus.FulfillmentDelivery, PlacedAt: now.Add(-15 * time.Minute),
		},
		TransitionOut: &orderrepo.Order{ID: orderID, Status: orderstatus.StatusCancelled},
	}
	uc := expire.New(orders, uctest.FixedNower{T: now}, 10*time.Minute)

	out, err := uc.Run(context.Background(), expire.In{OrderID: orderID})
	require.NoError(t, err)
	assert.True(t, out.Processed)
}

func TestRun_OrderNotFound(t *testing.T) {
	orders := &uctest.OrderRepo{GetByIDErr: pgerr.ErrNotFound}
	uc := expire.New(orders, uctest.FixedNower{T: time.Now()}, 10*time.Minute)

	out, err := uc.Run(context.Background(), expire.In{OrderID: uuid.New()})
	require.NoError(t, err)
	assert.False(t, out.Processed)
}

func TestRun_NotPlaced(t *testing.T) {
	orderID := uuid.New()
	orders := &uctest.OrderRepo{
		Order: &orderrepo.Order{
			ID: orderID, Status: orderstatus.StatusAccepted,
			FulfillmentType: orderstatus.FulfillmentDelivery, PlacedAt: time.Now(),
		},
	}
	uc := expire.New(orders, uctest.FixedNower{T: time.Now()}, 10*time.Minute)

	out, err := uc.Run(context.Background(), expire.In{OrderID: orderID})
	require.NoError(t, err)
	assert.False(t, out.Processed)
}

func TestRun_TransitionRace(t *testing.T) {
	orderID := uuid.New()
	orders := &uctest.OrderRepo{
		Order: &orderrepo.Order{
			ID: orderID, Status: orderstatus.StatusPlaced,
			FulfillmentType: orderstatus.FulfillmentDelivery, PlacedAt: time.Now(),
		},
		TransitionErr: pgerr.ErrNotFound,
	}
	uc := expire.New(orders, uctest.FixedNower{T: time.Now()}, 10*time.Minute)

	out, err := uc.Run(context.Background(), expire.In{OrderID: orderID})
	require.NoError(t, err)
	assert.False(t, out.Processed)
}
