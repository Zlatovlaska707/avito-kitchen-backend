package httpx

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type APIError struct {
	Status  int
	Code    string
	Message string
	Details map[string]any
}

type PageInfo struct {
	HasNext bool    `json:"has_next"`
	Cursor  *string `json:"cursor,omitempty"`
	Limit   int     `json:"limit"`
}

type cursorPayload struct {
	TS string    `json:"ts"`
	ID uuid.UUID `json:"id"`
}

func RespondJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if body == nil || status == http.StatusNoContent {
		return
	}
	_ = json.NewEncoder(w).Encode(body)
}

func RespondError(w http.ResponseWriter, ctx context.Context, apiErr APIError, _ error) {
	requestID := RequestIDFromContext(ctx)
	resp := map[string]any{
		"error": map[string]any{
			"code":       apiErr.Code,
			"message":    apiErr.Message,
			"request_id": requestID,
		},
	}
	if len(apiErr.Details) > 0 {
		resp["error"].(map[string]any)["details"] = apiErr.Details
	}
	if apiErr.Status == http.StatusTooManyRequests || apiErr.Status == http.StatusServiceUnavailable {
		w.Header().Set("Retry-After", "5")
	}
	RespondJSON(w, apiErr.Status, resp)
}

func DecodeJSON(r *http.Request, dst any) error {
	if r.Body == nil {
		return errors.New("empty body")
	}
	defer func() { _ = r.Body.Close() }()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		if errors.Is(err, io.EOF) {
			return errors.New("empty body")
		}
		return err
	}
	if dec.More() {
		return errors.New("multiple json values")
	}
	return nil
}

func ParseLimit(r *http.Request, defaultLimit, maxLimit int) (int, error) {
	raw := r.URL.Query().Get("limit")
	if raw == "" {
		return defaultLimit, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return 0, errors.New("invalid limit")
	}
	if n > maxLimit {
		n = maxLimit
	}
	return n, nil
}

func ParseCursor(r *http.Request) (*time.Time, *uuid.UUID, error) {
	raw := strings.TrimSpace(r.URL.Query().Get("cursor"))
	if raw == "" {
		return nil, nil, nil
	}
	data, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return nil, nil, errors.New("invalid cursor")
	}
	var payload cursorPayload
	if unmarshalErr := json.Unmarshal(data, &payload); unmarshalErr != nil {
		return nil, nil, errors.New("invalid cursor")
	}
	ts, err := time.Parse(time.RFC3339Nano, payload.TS)
	if err != nil {
		return nil, nil, errors.New("invalid cursor")
	}
	return &ts, &payload.ID, nil
}

func EncodeCursor(ts time.Time, id uuid.UUID) string {
	payload, _ := json.Marshal(cursorPayload{TS: ts.UTC().Format(time.RFC3339Nano), ID: id})
	return base64.RawURLEncoding.EncodeToString(payload)
}

func BuildPageInfo(limit int, hasNext bool, lastTS *time.Time, lastID *uuid.UUID) PageInfo {
	info := PageInfo{HasNext: hasNext, Limit: limit}
	if hasNext && lastTS != nil && lastID != nil {
		c := EncodeCursor(*lastTS, *lastID)
		info.Cursor = &c
	}
	return info
}

type requestIDKey struct{}

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}

func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}
