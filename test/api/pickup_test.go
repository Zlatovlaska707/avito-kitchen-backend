//go:build e2e

package api_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const expectedPickupTotal = int64(118000) // 2×(49000+10000), no delivery fee

func createPickupOrderBody(total int64) map[string]any {
	return map[string]any{
		"fulfillment_type":     "pickup",
		"contact_phone":        "+79001234567",
		"payment_method":       "cash",
		"expected_total_minor": total,
	}
}

func TestHappyPath_PickupToCompleted(t *testing.T) {
	base := requireAPI(t)
	clearCart(t, base)
	addDefaultCartItem(t, base)

	resp := customerRequest(t, http.MethodPost, base+"/api/v1/orders", createPickupOrderBody(expectedPickupTotal), map[string]string{
		"Idempotency-Key": uniqueKey("e2e-pickup"),
	})
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created struct {
		Order struct {
			ID              string `json:"id"`
			Status          string `json:"status"`
			FulfillmentType string `json:"fulfillment_type"`
		} `json:"order"`
	}
	decodeJSONBody(t, resp, &created)
	require.NotEmpty(t, created.Order.ID)
	assert.Equal(t, "placed", created.Order.Status)
	assert.Equal(t, "pickup", created.Order.FulfillmentType)

	pollOrderStatus(t, base, created.Order.ID, "completed", 3*time.Minute)
}
