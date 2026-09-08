package get_for_customer_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zlatovlaska707/avito-kitchen/internal/domain/orderstatus"
	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/pgerr"
	deliveryrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/delivery"
	orderrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/order"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/ordering/get_for_customer"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/uctest"
)

func TestRun_HappyPath(t *testing.T) {
	orderID := uuid.New()
	customerID := uuid.New()
	deliveryID := uuid.New()

	orders := &uctest.OrderRepo{
		Order: &orderrepo.Order{
			ID: orderID, CustomerID: customerID, Status: orderstatus.StatusPlaced,
			FulfillmentType: orderstatus.FulfillmentDelivery,
		},
		Items:   []orderrepo.Item{{ID: uuid.New(), OrderID: orderID, NameSnapshot: "Pizza", Quantity: 1}},
		History: []orderrepo.StatusEvent{{ToStatus: orderstatus.StatusPlaced, ActorType: orderstatus.ActorCustomer}},
	}
	deliveries := &uctest.DeliveryRepo{
		Delivery: &deliveryrepo.Delivery{ID: deliveryID, OrderID: orderID, Status: "pending"},
	}
	uc := get_for_customer.New(orders, deliveries)

	out, err := uc.Run(context.Background(), get_for_customer.In{OrderID: orderID, CustomerID: customerID})
	require.NoError(t, err)
	require.NotNil(t, out.Order)
	assert.Equal(t, orderID, out.Order.ID)
	require.Len(t, out.Items, 1)
	require.Len(t, out.History, 1)
	assert.Equal(t, orderstatus.StatusPlaced, out.History[0].ToStatus)
	require.NotNil(t, out.Delivery)
	assert.Equal(t, deliveryID, out.Delivery.ID)
}

func TestRun_GetHistoryError(t *testing.T) {
	orderID := uuid.New()
	customerID := uuid.New()
	orders := &uctest.OrderRepo{
		Order:      &orderrepo.Order{ID: orderID, CustomerID: customerID, Status: orderstatus.StatusPlaced},
		HistoryErr: assert.AnError,
	}
	uc := get_for_customer.New(orders, &uctest.DeliveryRepo{})

	_, err := uc.Run(context.Background(), get_for_customer.In{OrderID: orderID, CustomerID: customerID})
	assert.ErrorIs(t, err, assert.AnError)
}

func TestRun_OrderNotFound(t *testing.T) {
	orders := &uctest.OrderRepo{GetByCustomerErr: pgerr.ErrNotFound}
	uc := get_for_customer.New(orders, &uctest.DeliveryRepo{})

	_, err := uc.Run(context.Background(), get_for_customer.In{OrderID: uuid.New(), CustomerID: uuid.New()})
	assert.ErrorIs(t, err, usecase.ErrOrderNotFound)
}

func TestRun_GetItemsError(t *testing.T) {
	orderID := uuid.New()
	customerID := uuid.New()
	orders := &uctest.OrderRepo{
		Order: &orderrepo.Order{ID: orderID, CustomerID: customerID, Status: orderstatus.StatusPlaced},
		ItemsErr: assert.AnError,
	}
	uc := get_for_customer.New(orders, &uctest.DeliveryRepo{})

	_, err := uc.Run(context.Background(), get_for_customer.In{OrderID: orderID, CustomerID: customerID})
	assert.ErrorIs(t, err, assert.AnError)
}
