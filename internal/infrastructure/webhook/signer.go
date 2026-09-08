package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
)

func Sign(secret string, timestamp int64, body []byte) string {
	payload := strconv.FormatInt(timestamp, 10) + "." + string(body)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func Verify(secret string, timestamp int64, body []byte, signature string, now int64, maxSkew int64) error {
	if signature == "" {
		return fmt.Errorf("missing signature")
	}
	if abs(now-timestamp) > maxSkew {
		return fmt.Errorf("timestamp skew too large")
	}
	expected := Sign(secret, timestamp, body)
	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return fmt.Errorf("invalid signature")
	}
	return nil
}

func abs(v int64) int64 {
	if v < 0 {
		return -v
	}
	return v
}
