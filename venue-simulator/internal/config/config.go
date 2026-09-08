package config

import (
	"os"
	"time"
)

type Config struct {
	PlatformURL      string
	APIKey           string
	WebhookPublicURL string
	HTTPAddress      string
	MenuFile         string
	Mode             string
	AcceptDelay      time.Duration
	CookDelay        time.Duration
	HandoverDelay    time.Duration
	PollingEnabled   bool
	PollInterval     time.Duration
	WebhookSecret    string
}

func Load() Config {
	return Config{
		PlatformURL:      getenv("PLATFORM_API_URL", getenv("PLATFORM_URL", "http://api:8080")),
		APIKey:           os.Getenv("PARTNER_API_KEY"),
		WebhookPublicURL: getenv("WEBHOOK_PUBLIC_URL", "http://venue-simulator:8081/webhooks/orders"),
		HTTPAddress:      getenv("HTTP_ADDRESS", ":8081"),
		MenuFile:         getenv("MENU_FILE", "./menu.json"),
		Mode:             getenv("MODE", "auto_accept"),
		AcceptDelay:      durationEnv("ACCEPT_DELAY", 3*time.Second),
		CookDelay:        durationEnv("COOK_DELAY", 10*time.Second),
		HandoverDelay:    durationEnv("HANDOVER_DELAY", 5*time.Second),
		PollingEnabled:   os.Getenv("POLLING_ENABLED") != "false",
		PollInterval:     durationEnv("POLL_INTERVAL", 5*time.Second),
		WebhookSecret:    getenv("WEBHOOK_SECRET", "local-simulator-webhook-secret-change-me"),
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func durationEnv(k string, def time.Duration) time.Duration {
	if v := os.Getenv(k); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
