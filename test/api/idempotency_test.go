//go:build e2e

package api_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIdempotency_DoublePOST(t *testing.T) {
	base := requireAPI(t)
	clearCart(t, base)
	idemKey := "e2e-idempotency-" + time.Now().Format("20060102150405")

	addDefaultCartItem(t, base)
	body := createOrderBody(expectedTotal)

	resp1 := customerRequest(t, http.MethodPost, base+"/api/v1/orders", body, map[string]string{
		"Idempotency-Key": idemKey,
	})
	require.Equal(t, http.StatusCreated, resp1.StatusCode)

	var first struct {
		Order struct {
			ID string `json:"id"`
		} `json:"order"`
	}
	decodeJSONBody(t, resp1, &first)
	require.NotEmpty(t, first.Order.ID)

	resp2 := customerRequest(t, http.MethodPost, base+"/api/v1/orders", body, map[string]string{
		"Idempotency-Key": idemKey,
	})
	require.Equal(t, http.StatusCreated, resp2.StatusCode)
	assert.Equal(t, "true", resp2.Header.Get("Idempotency-Replayed"))

	var second struct {
		Order struct {
			ID string `json:"id"`
		} `json:"order"`
	}
	decodeJSONBody(t, resp2, &second)
	assert.Equal(t, first.Order.ID, second.Order.ID)
}
