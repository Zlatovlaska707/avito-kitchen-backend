//go:build e2e

package api_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHappyPath_DeliveryToCompleted(t *testing.T) {
	base := requireAPI(t)
	clearCart(t, base)
	addDefaultCartItem(t, base)

	resp := customerRequest(t, http.MethodPost, base+"/api/v1/orders", createOrderBody(expectedTotal), map[string]string{
		"Idempotency-Key": "e2e-happy-path-" + time.Now().Format("20060102150405"),
	})
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created struct {
		Order struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"order"`
	}
	decodeJSONBody(t, resp, &created)
	require.NotEmpty(t, created.Order.ID)
	assert.Equal(t, "placed", created.Order.Status)

	pollOrderStatus(t, base, created.Order.ID, "completed", 3*time.Minute)
}
