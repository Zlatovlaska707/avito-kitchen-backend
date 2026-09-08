package platformapi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// NewPartnerClient creates an oapi-codegen client with Bearer partner auth.
func NewPartnerClient(platformURL, apiKey string, timeout time.Duration) (*ClientWithResponses, error) {
	return NewClientWithResponses(platformURL,
		WithHTTPClient(&http.Client{Timeout: timeout}),
		WithRequestEditorFn(func(_ context.Context, req *http.Request) error {
			req.Header.Set("Authorization", "Bearer "+apiKey)
			return nil
		}),
	)
}

// HTTPError describes a non-success API response.
type HTTPError struct {
	Status int
	Body   string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("platform api status %d: %s", e.Status, e.Body)
}

// CheckStatus returns nil for 2xx responses.
func CheckStatus(status int, body []byte) error {
	if status >= 200 && status < 300 {
		return nil
	}
	return &HTTPError{Status: status, Body: string(body)}
}

// ReadBody reads and closes an HTTP response body.
func ReadBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
