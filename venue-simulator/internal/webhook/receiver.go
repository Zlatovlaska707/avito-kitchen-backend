package webhook

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/zlatovlaska707/avito-kitchen/venue-simulator/internal/config"
	"github.com/zlatovlaska707/avito-kitchen/venue-simulator/internal/processor"
	"github.com/zlatovlaska707/avito-kitchen/venue-simulator/internal/runtime"
)

const signatureMaxSkew = 300

type Receiver struct {
	mu        sync.Mutex
	orders    map[string]json.RawMessage
	processor *processor.OrderProcessor
	cfg       config.Config
	state     *runtime.State
}

func NewReceiver(cfg config.Config, state *runtime.State) (*Receiver, error) {
	proc, err := processor.New(cfg, state)
	if err != nil {
		return nil, err
	}
	return &Receiver{
		orders:    make(map[string]json.RawMessage),
		processor: proc,
		cfg:       cfg,
		state:     state,
	}, nil
}

func (r *Receiver) HandleOrderCreated(w http.ResponseWriter, req *http.Request) {
	if r.state.Mode() == "broken_webhook" {
		http.Error(w, "broken", http.StatusInternalServerError)
		return
	}
	body, err := io.ReadAll(io.LimitReader(req.Body, 1<<20))
	if err != nil {
		http.Error(w, "bad body", http.StatusBadRequest)
		return
	}
	if err := r.verifySignature(req, body); err != nil {
		log.Printf("webhook signature rejected: %v", err)
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}
	deliveryID := req.Header.Get("X-Avito-Kitchen-Delivery-Id")
	if !r.state.MarkDelivery(deliveryID) {
		w.WriteHeader(http.StatusOK)
		return
	}
	orderID := extractOrderID(body)
	r.mu.Lock()
	r.orders[orderID] = body
	r.mu.Unlock()
	log.Printf("webhook received order event order_id=%s delivery_id=%s", orderID, deliveryID)
	if orderID != "" {
		go r.processor.ProcessOrder(orderID, body)
	}
	w.WriteHeader(http.StatusOK)
}

func (r *Receiver) verifySignature(req *http.Request, body []byte) error {
	secret := r.state.WebhookSecret()
	if secret == "" {
		return nil
	}
	sig := req.Header.Get("X-Avito-Kitchen-Signature")
	tsRaw := req.Header.Get("X-Avito-Kitchen-Timestamp")
	ts, err := strconv.ParseInt(tsRaw, 10, 64)
	if err != nil {
		return err
	}
	return verify(secret, ts, body, sig, time.Now().Unix(), signatureMaxSkew)
}

func (r *Receiver) ProcessPolledOrder(orderID string) {
	r.mu.Lock()
	if _, seen := r.orders[orderID]; seen {
		r.mu.Unlock()
		return
	}
	r.orders[orderID] = nil
	r.mu.Unlock()
	log.Printf("poll processing order %s", orderID)
	go r.processor.ProcessOrder(orderID, nil)
}

func (r *Receiver) HandleHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func (r *Receiver) HandleOrders(w http.ResponseWriter, _ *http.Request) {
	r.mu.Lock()
	defer r.mu.Unlock()
	ids := make([]string, 0, len(r.orders))
	for id := range r.orders {
		if id != "" {
			ids = append(ids, id)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"order_ids": ids})
}

func extractOrderID(body []byte) string {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return ""
	}
	if orderID, _ := payload["order_id"].(string); orderID != "" {
		return orderID
	}
	if order, ok := payload["order"].(map[string]any); ok {
		orderID, _ := order["id"].(string)
		return orderID
	}
	return ""
}
