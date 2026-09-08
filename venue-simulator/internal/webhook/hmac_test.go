package webhook

import "testing"

func TestVerify_GoldenVector(t *testing.T) {
	secret := "test-webhook-secret"
	ts := int64(1_700_000_000)
	body := []byte(`{"order_id":"11111111-1111-1111-1111-111111111111","event_type":"order.created"}`)
	sig := "sha256=1f66afdbbe103928d484de9d141057ad47c1e1b9012aa632adf99988e2bec070"
	if err := verify(secret, ts, body, sig, ts, 300); err != nil {
		t.Fatalf("verify: %v", err)
	}
	if got := sign(secret, ts, body); got != sig {
		t.Fatalf("sign mismatch: %s", got)
	}
}
