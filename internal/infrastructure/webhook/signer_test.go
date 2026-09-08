package webhook

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignVerify_GoldenVector(t *testing.T) {
	secret := "test-webhook-secret"
	ts := int64(1_700_000_000)
	body := []byte(`{"order_id":"11111111-1111-1111-1111-111111111111","event_type":"order.created"}`)
	sig := Sign(secret, ts, body)
	assert.Equal(t, "sha256=1f66afdbbe103928d484de9d141057ad47c1e1b9012aa632adf99988e2bec070", sig)
	require.NoError(t, Verify(secret, ts, body, sig, ts, 300))
}

func TestVerify_RejectsInvalidSignature(t *testing.T) {
	err := Verify("secret", 1_700_000_000, []byte("{}"), "sha256=deadbeef", 1_700_000_000, 300)
	assert.Error(t, err)
}

func TestVerify_RejectsTimestampSkew(t *testing.T) {
	secret := "secret"
	ts := int64(1_700_000_000)
	body := []byte("{}")
	sig := Sign(secret, ts, body)
	err := Verify(secret, ts, body, sig, ts+400, 300)
	assert.Error(t, err)
}
