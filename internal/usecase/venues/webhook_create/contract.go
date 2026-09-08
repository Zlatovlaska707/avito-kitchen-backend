package webhook_create

import (
	"github.com/google/uuid"
	webhookrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/webhook"
)

type In struct {
	VenueID    uuid.UUID
	URL        string
	Secret     string
	EventTypes []string
}

type Out struct {
	Webhook webhookrepo.Webhook
}
