package runtime

import (
	"sync"

	"github.com/zlatovlaska707/avito-kitchen/venue-simulator/internal/config"
)

const defaultWebhookSecret = "local-simulator-webhook-secret-change-me"

// State holds simulator runtime configuration that E2E tests can change via Control API.
type State struct {
	mu            sync.RWMutex
	mode          string
	webhookSecret string
	deliverySeen  map[string]struct{}
}

func NewState(cfg config.Config) *State {
	secret := cfg.WebhookSecret
	if secret == "" {
		secret = defaultWebhookSecret
	}
	return &State{
		mode:          cfg.Mode,
		webhookSecret: secret,
		deliverySeen:  make(map[string]struct{}),
	}
}

func (s *State) Mode() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.mode
}

func (s *State) SetMode(mode string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mode = mode
}

func (s *State) WebhookSecret() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.webhookSecret
}

func (s *State) SetWebhookSecret(secret string) {
	if secret == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.webhookSecret = secret
}

func (s *State) MarkDelivery(id string) bool {
	if id == "" {
		return true
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.deliverySeen[id]; ok {
		return false
	}
	s.deliverySeen[id] = struct{}{}
	return true
}
