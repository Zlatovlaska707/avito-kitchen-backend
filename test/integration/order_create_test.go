//go:build integration

package integration_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/cart/item_add"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/ordering/create"
	"github.com/zlatovlaska707/avito-kitchen/test/suite"
)

func TestOrderCreate_Atomicity(t *testing.T) {
	pool := suite.Pool(t)
	suite.Reset(t, pool)
	fx := suite.NewFixtures(t, pool)
	ctx := context.Background()

	_, err := fx.CartAdd.Run(ctx, item_add.In{
		CustomerID: suite.CustomerID,
		VenueID:    suite.VenuePizzaMario,
		MenuItemID: suite.MenuItemMargherita,
		Quantity:   2,
		OptionIDs:  []uuid.UUID{suite.SizeMedium},
	})
	require.NoError(t, err)

	addr := deliveryAddress(55.7558, 37.6176)
	out, err := fx.OrderCreate.Run(ctx, create.In{
		CustomerID:      suite.CustomerID,
		FulfillmentType: "delivery",
		DeliveryAddress: addr,
		ContactPhone:    "+79001234567",
		PaymentMethod:   "cash",
		IdempotencyKey:  "atomicity-key-1",
		RequestHash:     "hash-atomicity-1",
	})
	require.NoError(t, err)
	require.NotNil(t, out.Order)
	orderID := out.Order.ID

	var orderCount, itemCount, historyCount, outboxCount, deliveryCount int
	var cartStatus string

	err = pool.QueryRow(ctx, `SELECT COUNT(*) FROM orders WHERE id = $1`, orderID).Scan(&orderCount)
	require.NoError(t, err)
	err = pool.QueryRow(ctx, `SELECT COUNT(*) FROM order_items WHERE order_id = $1`, orderID).Scan(&itemCount)
	require.NoError(t, err)
	err = pool.QueryRow(ctx, `SELECT COUNT(*) FROM order_status_history WHERE order_id = $1`, orderID).Scan(&historyCount)
	require.NoError(t, err)
	err = pool.QueryRow(ctx, `SELECT COUNT(*) FROM outbox_events WHERE aggregate_id = $1`, orderID).Scan(&outboxCount)
	require.NoError(t, err)
	err = pool.QueryRow(ctx, `SELECT COUNT(*) FROM deliveries WHERE order_id = $1`, orderID).Scan(&deliveryCount)
	require.NoError(t, err)
	err = pool.QueryRow(ctx, `SELECT status FROM carts WHERE customer_id = $1`, suite.CustomerID).Scan(&cartStatus)
	require.NoError(t, err)

	assert.Equal(t, 1, orderCount)
	assert.Equal(t, 1, itemCount)
	assert.GreaterOrEqual(t, historyCount, 1)
	assert.GreaterOrEqual(t, outboxCount, 1)
	assert.Equal(t, 1, deliveryCount)
	assert.Equal(t, "converted", cartStatus)
}

func TestOrderCreate_ConcurrentIdempotency(t *testing.T) {
	pool := suite.Pool(t)
	suite.Reset(t, pool)
	fx := suite.NewFixtures(t, pool)
	ctx := context.Background()

	_, err := fx.CartAdd.Run(ctx, item_add.In{
		CustomerID: suite.CustomerID,
		VenueID:    suite.VenuePizzaMario,
		MenuItemID: suite.MenuItemMargherita,
		Quantity:   2,
		OptionIDs:  []uuid.UUID{suite.SizeMedium},
	})
	require.NoError(t, err)

	const workers = 10
	idemKey := "concurrent-idem-key"
	reqHash := "concurrent-idem-hash"
	addr := deliveryAddress(55.7558, 37.6176)

	var wg sync.WaitGroup
	results := make([]*create.Out, workers)
	errs := make([]error, workers)

	for i := range workers {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			out, runErr := fx.OrderCreate.Run(ctx, create.In{
				CustomerID:      suite.CustomerID,
				FulfillmentType: "delivery",
				DeliveryAddress: addr,
				ContactPhone:    "+79001234567",
				PaymentMethod:   "cash",
				IdempotencyKey:  idemKey,
				RequestHash:     reqHash,
			})
			results[idx] = out
			errs[idx] = runErr
		}(i)
	}
	wg.Wait()

	var successOrderID uuid.UUID
	successCount := 0
	for i, runErr := range errs {
		if runErr == nil {
			require.NotNil(t, results[i].Order)
			successCount++
			if successOrderID == uuid.Nil {
				successOrderID = results[i].Order.ID
			} else {
				assert.Equal(t, successOrderID, results[i].Order.ID)
			}
			continue
		}
		assert.True(t,
			errors.Is(runErr, usecase.ErrRequestInProgress) ||
				errors.Is(runErr, usecase.ErrCartNotFound) ||
				errors.Is(runErr, usecase.ErrIdempotencyKeyReuse),
			"worker %d unexpected error: %v", i, runErr,
		)
	}
	assert.Equal(t, 1, successCount)

	var orderCount int
	err = pool.QueryRow(ctx, `SELECT COUNT(*) FROM orders WHERE customer_id = $1`, suite.CustomerID).Scan(&orderCount)
	require.NoError(t, err)
	assert.Equal(t, 1, orderCount)
}

func TestOrderCreate_ConcurrentCart(t *testing.T) {
	pool := suite.Pool(t)
	suite.Reset(t, pool)
	fx := suite.NewFixtures(t, pool)
	ctx := context.Background()

	_, err := fx.CartAdd.Run(ctx, item_add.In{
		CustomerID: suite.CustomerID,
		VenueID:    suite.VenuePizzaMario,
		MenuItemID: suite.MenuItemMargherita,
		Quantity:   2,
		OptionIDs:  []uuid.UUID{suite.SizeMedium},
	})
	require.NoError(t, err)

	const workers = 10
	addr := deliveryAddress(55.7558, 37.6176)

	var wg sync.WaitGroup
	errs := make([]error, workers)

	for i := range workers {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, runErr := fx.OrderCreate.Run(ctx, create.In{
				CustomerID:      suite.CustomerID,
				FulfillmentType: "delivery",
				DeliveryAddress: addr,
				ContactPhone:    "+79001234567",
				PaymentMethod:   "cash",
				IdempotencyKey:  fmt.Sprintf("cart-race-key-%d", idx),
				RequestHash:     fmt.Sprintf("cart-race-hash-%d", idx),
			})
			errs[idx] = runErr
		}(i)
	}
	wg.Wait()

	successCount := 0
	cartNotFoundCount := 0
	for _, runErr := range errs {
		if runErr == nil {
			successCount++
			continue
		}
		if assert.ErrorIs(t, runErr, usecase.ErrCartNotFound) {
			cartNotFoundCount++
		}
	}
	assert.Equal(t, 1, successCount)
	assert.Equal(t, workers-1, cartNotFoundCount)

	var orderCount int
	err = pool.QueryRow(ctx, `SELECT COUNT(*) FROM orders WHERE customer_id = $1`, suite.CustomerID).Scan(&orderCount)
	require.NoError(t, err)
	assert.Equal(t, 1, orderCount)
}

func deliveryAddress(lat, lon float64) json.RawMessage {
	raw, _ := json.Marshal(map[string]any{"text": "Moscow, Tverskaya 1", "lat": lat, "lon": lon})
	return raw
}
