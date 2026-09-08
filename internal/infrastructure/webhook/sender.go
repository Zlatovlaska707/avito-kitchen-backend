package webhook

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

type SendRequest struct {
	URL        string
	Secret     string
	EventType  string
	DeliveryID string
	Payload    []byte
	Timeout    time.Duration
}

type SendResult struct {
	StatusCode int
	Duration   time.Duration
	Body       []byte
}

type Sender struct {
	client *http.Client
}

func NewSender(dialGuard func(ctx context.Context, network, addr string) (net.Conn, error)) *Sender {
	transport := &http.Transport{
		DialContext: dialGuard,
	}
	return &Sender{client: &http.Client{
		Timeout:   0,
		Transport: transport,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}}
}

func (s *Sender) Send(ctx context.Context, req SendRequest) (*SendResult, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, req.URL, bytes.NewReader(req.Payload))
	if err != nil {
		return nil, fmt.Errorf("build webhook request: %w", err)
	}
	ts := time.Now().Unix()
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Avito-Kitchen-Event", req.EventType)
	httpReq.Header.Set("X-Avito-Kitchen-Delivery-Id", req.DeliveryID)
	httpReq.Header.Set("X-Avito-Kitchen-Timestamp", fmt.Sprintf("%d", ts))
	httpReq.Header.Set("X-Avito-Kitchen-Signature", Sign(req.Secret, ts, req.Payload))

	start := time.Now()
	client := s.client
	if req.Timeout > 0 {
		client = &http.Client{
			Timeout:       req.Timeout,
			Transport:     s.client.Transport,
			CheckRedirect: s.client.CheckRedirect,
		}
	}
	resp, err := client.Do(httpReq)
	duration := time.Since(start)
	if err != nil {
		return nil, fmt.Errorf("send webhook: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	return &SendResult{StatusCode: resp.StatusCode, Duration: duration, Body: body}, nil
}
