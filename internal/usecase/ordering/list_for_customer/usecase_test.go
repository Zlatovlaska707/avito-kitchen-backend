package list_for_customer_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zlatovlaska707/avito-kitchen/internal/domain/orderstatus"
	orderrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/order"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/ordering/list_for_customer"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/uctest"
)

func makeOrders(n int, customerID uuid.UUID) []orderrepo.Order {
	orders := make([]orderrepo.Order, n)
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	for i := range orders {
		orders[i] = orderrepo.Order{
			ID:         uuid.New(),
			CustomerID: customerID,
			Status:     orderstatus.StatusPlaced,
			CreatedAt:  base.Add(time.Duration(i) * time.Minute),
		}
	}
	return orders
}

func TestRun_HappyPath(t *testing.T) {
	customerID := uuid.New()
	orders := makeOrders(2, customerID)
	repo := &uctest.OrderRepo{ListOrders: orders}
	uc := list_for_customer.New(repo)

	out, err := uc.Run(context.Background(), list_for_customer.In{
		CustomerID: &customerID,
		Limit:      10,
	})
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Len(t, out.Orders, 2)
	assert.False(t, out.HasNext)
	require.NotNil(t, out.LastTS)
	require.NotNil(t, out.LastID)
	assert.Equal(t, orders[1].ID, *out.LastID)
}

func TestRun_DefaultLimit(t *testing.T) {
	customerID := uuid.New()
	orders := makeOrders(21, customerID)
	repo := &uctest.OrderRepo{ListOrders: orders}
	uc := list_for_customer.New(repo)

	out, err := uc.Run(context.Background(), list_for_customer.In{CustomerID: &customerID})
	require.NoError(t, err)
	assert.Len(t, out.Orders, 20)
	assert.True(t, out.HasNext)
}

func TestRun_HasNextPagination(t *testing.T) {
	customerID := uuid.New()
	orders := makeOrders(3, customerID)
	repo := &uctest.OrderRepo{ListOrders: orders}
	uc := list_for_customer.New(repo)

	out, err := uc.Run(context.Background(), list_for_customer.In{
		CustomerID: &customerID,
		Limit:      2,
	})
	require.NoError(t, err)
	assert.Len(t, out.Orders, 2)
	assert.True(t, out.HasNext)
	assert.Equal(t, orders[1].ID, *out.LastID)
}

func TestRun_EmptyResult(t *testing.T) {
	repo := &uctest.OrderRepo{ListOrders: nil}
	uc := list_for_customer.New(repo)

	out, err := uc.Run(context.Background(), list_for_customer.In{Limit: 10})
	require.NoError(t, err)
	assert.Empty(t, out.Orders)
	assert.False(t, out.HasNext)
	assert.Nil(t, out.LastTS)
	assert.Nil(t, out.LastID)
}

func TestRun_WithStatusFilter(t *testing.T) {
	customerID := uuid.New()
	status := string(orderstatus.StatusReady)
	repo := &uctest.OrderRepo{
		ListOrders: []orderrepo.Order{{ID: uuid.New(), CustomerID: customerID, Status: orderstatus.StatusReady}},
	}
	uc := list_for_customer.New(repo)

	out, err := uc.Run(context.Background(), list_for_customer.In{
		CustomerID: &customerID,
		Status:     &status,
		Limit:      10,
	})
	require.NoError(t, err)
	assert.Len(t, out.Orders, 1)
}

func TestRun_ListError(t *testing.T) {
	repo := &uctest.OrderRepo{ListErr: assert.AnError}
	uc := list_for_customer.New(repo)

	_, err := uc.Run(context.Background(), list_for_customer.In{Limit: 10})
	assert.ErrorIs(t, err, assert.AnError)
}
