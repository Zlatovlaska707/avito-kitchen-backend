package app

import (
	"log/slog"

	"github.com/zlatovlaska707/avito-kitchen/internal/config"
	cartrepo "github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/cart"
	deliveryrepo "github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/delivery"
	idempotencyrepo "github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/idempotency"
	orderrepo "github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/order"
	outboxrepo "github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/outbox"
	webhookrepo "github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/webhook"
	webhookinfra "github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/webhook"
	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/worker"
	accepttimeoutworker "github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/worker/accept_timeout"
	cartgcworker "github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/worker/cart_gc"
	deliverysimworker "github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/worker/delivery_simulator"
	idempotencygcworker "github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/worker/idempotency_gc"
	outboxworker "github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/worker/outbox_publisher"
	stoplistworker "github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/worker/stoplist_restore"
	catalogstoplist "github.com/zlatovlaska707/avito-kitchen/internal/usecase/catalog/stoplist_restore"
	deliveryadvance "github.com/zlatovlaska707/avito-kitchen/internal/usecase/delivery/advance"
	deliveryassign "github.com/zlatovlaska707/avito-kitchen/internal/usecase/delivery/assign"
	orderexpire "github.com/zlatovlaska707/avito-kitchen/internal/usecase/ordering/expire"
)

func setupWorkers(cfg config.Config, logger *slog.Logger, outbox *outboxrepo.Repository, webhooks *webhookrepo.Repository, orders *orderrepo.Repository, carts *cartrepo.Repository, idempotency *idempotencyrepo.Repository, deliveries *deliveryrepo.Repository, stoplist *catalogstoplist.UseCase, expire *orderexpire.UseCase, assign *deliveryassign.UseCase, advance *deliveryadvance.UseCase) *worker.Runner {
	if !cfg.Workers.Enabled {
		return nil
	}
	ssrf := webhookinfra.SSRFGuard{AllowPrivateIPs: cfg.Webhook.AllowPrivateIPs}
	sender := webhookinfra.NewSender(ssrf.DialContext)
	ws := []worker.Worker{
		outboxworker.New(outbox, webhooks, sender, cfg),
		accepttimeoutworker.New(orders, expire, cfg.Order.AcceptTimeout),
		stoplistworker.New(stoplist),
		cartgcworker.New(carts),
		idempotencygcworker.New(idempotency),
		deliverysimworker.New(deliveries, assign, advance, cfg.Delivery.AssignDelay, cfg.Delivery.TransitDelay, cfg.Delivery.SimulationEnabled),
	}
	return worker.NewRunner(logger, ws...)
}
