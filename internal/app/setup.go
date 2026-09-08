package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	trmmanager "github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zlatovlaska707/avito-kitchen/internal/config"
	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/nower"
	apikeyrepo "github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/apikey"
	cartrepo "github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/cart"
	catalogrepo "github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/catalog"
	deliveryrepo "github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/delivery"
	idempotencyrepo "github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/idempotency"
	orderrepo "github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/order"
	outboxrepo "github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/outbox"
	venuerepo "github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/venue"
	webhookrepo "github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/webhook"
	webhookinfra "github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/webhook"
	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/worker"

	cartclear "github.com/zlatovlaska707/avito-kitchen/internal/usecase/cart/clear"
	cartget "github.com/zlatovlaska707/avito-kitchen/internal/usecase/cart/get"
	cartitemadd "github.com/zlatovlaska707/avito-kitchen/internal/usecase/cart/item_add"
	cartitemupdate "github.com/zlatovlaska707/avito-kitchen/internal/usecase/cart/item_update"
	catalogavail "github.com/zlatovlaska707/avito-kitchen/internal/usecase/catalog/item_availability_set"
	catalogmenget "github.com/zlatovlaska707/avito-kitchen/internal/usecase/catalog/menu_get"
	catalogmenusync "github.com/zlatovlaska707/avito-kitchen/internal/usecase/catalog/menu_sync"
	catalogstoplist "github.com/zlatovlaska707/avito-kitchen/internal/usecase/catalog/stoplist_restore"
	deliveryadvance "github.com/zlatovlaska707/avito-kitchen/internal/usecase/delivery/advance"
	deliveryassign "github.com/zlatovlaska707/avito-kitchen/internal/usecase/delivery/assign"
	orderaccept "github.com/zlatovlaska707/avito-kitchen/internal/usecase/ordering/accept"
	ordercancel "github.com/zlatovlaska707/avito-kitchen/internal/usecase/ordering/cancel"
	ordercreate "github.com/zlatovlaska707/avito-kitchen/internal/usecase/ordering/create"
	orderexpire "github.com/zlatovlaska707/avito-kitchen/internal/usecase/ordering/expire"
	ordergetcustomer "github.com/zlatovlaska707/avito-kitchen/internal/usecase/ordering/get_for_customer"
	ordergetvenue "github.com/zlatovlaska707/avito-kitchen/internal/usecase/ordering/get_for_venue"
	orderhandover "github.com/zlatovlaska707/avito-kitchen/internal/usecase/ordering/handover"
	orderlistadmin "github.com/zlatovlaska707/avito-kitchen/internal/usecase/ordering/list_for_admin"
	orderlistcustomer "github.com/zlatovlaska707/avito-kitchen/internal/usecase/ordering/list_for_customer"
	orderlistvenue "github.com/zlatovlaska707/avito-kitchen/internal/usecase/ordering/list_for_venue"
	orderready "github.com/zlatovlaska707/avito-kitchen/internal/usecase/ordering/ready"
	orderreject "github.com/zlatovlaska707/avito-kitchen/internal/usecase/ordering/reject"
	venueapikeyissue "github.com/zlatovlaska707/avito-kitchen/internal/usecase/venues/apikey_issue"
	venueapikeyrevoke "github.com/zlatovlaska707/avito-kitchen/internal/usecase/venues/apikey_revoke"
	venueget "github.com/zlatovlaska707/avito-kitchen/internal/usecase/venues/get"
	venueslist "github.com/zlatovlaska707/avito-kitchen/internal/usecase/venues/list"
	venueonboard "github.com/zlatovlaska707/avito-kitchen/internal/usecase/venues/onboard"
	venuestatusset "github.com/zlatovlaska707/avito-kitchen/internal/usecase/venues/status_set"
	venueupdate "github.com/zlatovlaska707/avito-kitchen/internal/usecase/venues/update"
	webhookcreate "github.com/zlatovlaska707/avito-kitchen/internal/usecase/venues/webhook_create"
	webhookdelete "github.com/zlatovlaska707/avito-kitchen/internal/usecase/venues/webhook_delete"

	adminapikeyissue "github.com/zlatovlaska707/avito-kitchen/internal/handler/admin/apikey_issue"
	adminapikeyrevoke "github.com/zlatovlaska707/avito-kitchen/internal/handler/admin/apikey_revoke"
	adminorderlist "github.com/zlatovlaska707/avito-kitchen/internal/handler/admin/order_list"
	adminvenuecreate "github.com/zlatovlaska707/avito-kitchen/internal/handler/admin/venue_create"
	adminvenueupdate "github.com/zlatovlaska707/avito-kitchen/internal/handler/admin/venue_update"
	customercartclear "github.com/zlatovlaska707/avito-kitchen/internal/handler/customer/cart_clear"
	customercartget "github.com/zlatovlaska707/avito-kitchen/internal/handler/customer/cart_get"
	customercartitemadd "github.com/zlatovlaska707/avito-kitchen/internal/handler/customer/cart_item_add"
	customercartitemupdate "github.com/zlatovlaska707/avito-kitchen/internal/handler/customer/cart_item_update"
	customermenget "github.com/zlatovlaska707/avito-kitchen/internal/handler/customer/menu_get"
	customerordercancel "github.com/zlatovlaska707/avito-kitchen/internal/handler/customer/order_cancel"
	customerordercreate "github.com/zlatovlaska707/avito-kitchen/internal/handler/customer/order_create"
	customerorderget "github.com/zlatovlaska707/avito-kitchen/internal/handler/customer/order_get"
	customerorderlist "github.com/zlatovlaska707/avito-kitchen/internal/handler/customer/order_list"
	customervenueget "github.com/zlatovlaska707/avito-kitchen/internal/handler/customer/venue_get"
	customervenueslist "github.com/zlatovlaska707/avito-kitchen/internal/handler/customer/venues_list"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/middleware"
	partneritemavail "github.com/zlatovlaska707/avito-kitchen/internal/handler/partner/item_availability_set"
	partnermenusync "github.com/zlatovlaska707/avito-kitchen/internal/handler/partner/menu_sync"
	partnerorderaccept "github.com/zlatovlaska707/avito-kitchen/internal/handler/partner/order_accept"
	partnerorderget "github.com/zlatovlaska707/avito-kitchen/internal/handler/partner/order_get"
	partnerorderhandover "github.com/zlatovlaska707/avito-kitchen/internal/handler/partner/order_handover"
	partnerorderlist "github.com/zlatovlaska707/avito-kitchen/internal/handler/partner/order_list"
	partnerorderready "github.com/zlatovlaska707/avito-kitchen/internal/handler/partner/order_ready"
	partnerorderreject "github.com/zlatovlaska707/avito-kitchen/internal/handler/partner/order_reject"
	partnervenuestatus "github.com/zlatovlaska707/avito-kitchen/internal/handler/partner/venue_status_set"
	partnerwebhookcreate "github.com/zlatovlaska707/avito-kitchen/internal/handler/partner/webhook_create"
	partnerwebhookdelete "github.com/zlatovlaska707/avito-kitchen/internal/handler/partner/webhook_delete"
)

type Dependencies struct {
	Config   config.Config
	Pool     *pgxpool.Pool
	Logger   *slog.Logger
	Validate *validator.Validate

	Handlers *Handlers
	Workers  *worker.Runner
	OpenAPI  *middleware.OpenAPIValidator
}

type Handlers struct {
	CustomerVenuesList      *customervenueslist.Handler
	CustomerVenueGet        *customervenueget.Handler
	CustomerMenuGet         *customermenget.Handler
	CustomerCartGet         *customercartget.Handler
	CustomerCartItemAdd     *customercartitemadd.Handler
	CustomerCartItemUpdate  *customercartitemupdate.Handler
	CustomerCartClear       *customercartclear.Handler
	CustomerOrderCreate     *customerordercreate.Handler
	CustomerOrderList       *customerorderlist.Handler
	CustomerOrderGet        *customerorderget.Handler
	CustomerOrderCancel     *customerordercancel.Handler
	PartnerMenuSync         *partnermenusync.Handler
	PartnerItemAvailability *partneritemavail.Handler
	PartnerVenueStatusSet   *partnervenuestatus.Handler
	PartnerOrderList        *partnerorderlist.Handler
	PartnerOrderGet         *partnerorderget.Handler
	PartnerOrderAccept      *partnerorderaccept.Handler
	PartnerOrderReject      *partnerorderreject.Handler
	PartnerOrderReady       *partnerorderready.Handler
	PartnerOrderHandover    *partnerorderhandover.Handler
	PartnerWebhookCreate    *partnerwebhookcreate.Handler
	PartnerWebhookDelete    *partnerwebhookdelete.Handler
	AdminVenueCreate        *adminvenuecreate.Handler
	AdminVenueUpdate        *adminvenueupdate.Handler
	AdminAPIKeyIssue        *adminapikeyissue.Handler
	AdminAPIKeyRevoke       *adminapikeyrevoke.Handler
	AdminOrderList          *adminorderlist.Handler

	APIKeyRepo *apikeyrepo.CachedRepository
}

func setup(ctx context.Context, cfg config.Config, pool *pgxpool.Pool, logger *slog.Logger) (*Dependencies, error) {
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	getter := trmpgx.DefaultCtxGetter
	trManager := trmmanager.Must(trmpgx.NewDefaultFactory(pool))
	defTZ := cfg.DefaultLocation()
	now := nower.NewReal()
	v := validator.New()

	venues := venuerepo.New(pool, getter)
	catalog := catalogrepo.New(pool, getter)
	carts := cartrepo.New(pool, getter)
	orders := orderrepo.New(pool, getter)
	idempotency := idempotencyrepo.New(pool, getter)
	outbox := outboxrepo.New(pool, getter)
	webhooks := webhookrepo.New(pool, getter)
	deliveries := deliveryrepo.New(pool, getter)
	apikeys := apikeyrepo.New(pool, getter)
	cachedAPIKeys := apikeyrepo.NewCachedRepository(apikeys, 30*time.Second)

	ssrf := webhookinfra.SSRFGuard{AllowPrivateIPs: cfg.Webhook.AllowPrivateIPs}
	_ = webhookinfra.NewSender(ssrf.DialContext)

	// use cases
	ucVenuesList := venueslist.New(venues, now, defTZ)
	ucVenueGet := venueget.New(venues, now, defTZ)
	ucMenuGet := catalogmenget.New(catalog, venues)
	ucCartGet := cartget.New(carts, catalog, venues, now, defTZ)
	ucCartAdd := cartitemadd.New(carts, catalog, venues, now, defTZ, cfg.Cart.TTL, cfg.Cart.MaxItems)
	ucCartUpdate := cartitemupdate.New(carts, catalog, venues, now, defTZ)
	ucCartClear := cartclear.New(carts)
	ucOrderCreate := ordercreate.New(trManager, carts, catalog, venues, orders, idempotency, now, defTZ, cfg.Idempotency.TTL, cfg.Delivery.BaseMinutes, cfg.Delivery.MinutesPerKm)
	ucOrderListCustomer := orderlistcustomer.New(orders)
	ucOrderGetCustomer := ordergetcustomer.New(orders, deliveries)
	ucOrderCancel := ordercancel.New(orders, now, cfg.Order.CancelGrace, cfg.Order.AcceptTimeout)
	ucMenuSync := catalogmenusync.New(catalog, cfg.MenuSync.MaxCategories, cfg.MenuSync.MaxItems)
	ucItemAvail := catalogavail.New(catalog)
	ucVenueStatus := venuestatusset.New(venues)
	ucOrderListVenue := orderlistvenue.New(orders)
	ucOrderGetVenue := ordergetvenue.New(orders, deliveries)
	ucOrderAccept := orderaccept.New(orders, now, cfg.Order.CancelGrace, cfg.Order.AcceptTimeout)
	ucOrderReject := orderreject.New(orders, now, cfg.Order.CancelGrace, cfg.Order.AcceptTimeout)
	ucOrderReady := orderready.New(orders, now, cfg.Order.CancelGrace, cfg.Order.AcceptTimeout)
	ucOrderHandover := orderhandover.New(orders, now, cfg.Order.CancelGrace, cfg.Order.AcceptTimeout)
	ucWebhookCreate := webhookcreate.New(webhooks, &ssrf)
	ucWebhookDelete := webhookdelete.New(webhooks)
	ucVenueOnboard := venueonboard.New(venues)
	ucVenueUpdate := venueupdate.New(venues)
	ucAPIKeyIssue := venueapikeyissue.New(venues, apikeys)
	ucAPIKeyRevoke := venueapikeyrevoke.New(cachedAPIKeys)
	ucOrderListAdmin := orderlistadmin.New(orders)

	ucStoplistRestore := catalogstoplist.New(catalog)
	ucOrderExpire := orderexpire.New(orders, now, cfg.Order.AcceptTimeout)
	ucDeliveryAssign := deliveryassign.New(deliveries, orders)
	ucDeliveryAdvance := deliveryadvance.New(deliveries, orders)

	h := &Handlers{
		CustomerVenuesList:      customervenueslist.New(ucVenuesList, v),
		CustomerVenueGet:        customervenueget.New(ucVenueGet, v),
		CustomerMenuGet:         customermenget.New(ucMenuGet, v),
		CustomerCartGet:         customercartget.New(ucCartGet, v),
		CustomerCartItemAdd:     customercartitemadd.New(ucCartAdd, v),
		CustomerCartItemUpdate:  customercartitemupdate.New(ucCartUpdate, v),
		CustomerCartClear:       customercartclear.New(ucCartClear, v),
		CustomerOrderCreate:     customerordercreate.New(ucOrderCreate, v, idempotency, logger),
		CustomerOrderList:       customerorderlist.New(ucOrderListCustomer, v),
		CustomerOrderGet:        customerorderget.New(ucOrderGetCustomer, v),
		CustomerOrderCancel:     customerordercancel.New(ucOrderCancel, v),
		PartnerMenuSync:         partnermenusync.New(ucMenuSync, v),
		PartnerItemAvailability: partneritemavail.New(ucItemAvail, v),
		PartnerVenueStatusSet:   partnervenuestatus.New(ucVenueStatus, v),
		PartnerOrderList:        partnerorderlist.New(ucOrderListVenue, v),
		PartnerOrderGet:         partnerorderget.New(ucOrderGetVenue, v),
		PartnerOrderAccept:      partnerorderaccept.New(ucOrderAccept, v),
		PartnerOrderReject:      partnerorderreject.New(ucOrderReject, v),
		PartnerOrderReady:       partnerorderready.New(ucOrderReady, v),
		PartnerOrderHandover:    partnerorderhandover.New(ucOrderHandover, v),
		PartnerWebhookCreate:    partnerwebhookcreate.New(ucWebhookCreate, v),
		PartnerWebhookDelete:    partnerwebhookdelete.New(ucWebhookDelete, v),
		AdminVenueCreate:        adminvenuecreate.New(ucVenueOnboard, v),
		AdminVenueUpdate:        adminvenueupdate.New(ucVenueUpdate, v),
		AdminAPIKeyIssue:        adminapikeyissue.New(ucAPIKeyIssue, v),
		AdminAPIKeyRevoke:       adminapikeyrevoke.New(ucAPIKeyRevoke, v),
		AdminOrderList:          adminorderlist.New(ucOrderListAdmin, v),
		APIKeyRepo:              cachedAPIKeys,
	}

	var workers *worker.Runner
	if cfg.Workers.Enabled {
		workers = setupWorkers(cfg, logger, outbox, webhooks, orders, carts, idempotency, deliveries, ucStoplistRestore, ucOrderExpire, ucDeliveryAssign, ucDeliveryAdvance)
	}

	var openAPI *middleware.OpenAPIValidator
	if cfg.OpenAPI.ValidateRequests || cfg.OpenAPI.ValidateResponses {
		var oerr error
		openAPI, oerr = middleware.NewOpenAPIValidator(cfg.OpenAPI.ValidateRequests, cfg.OpenAPI.ValidateResponses, logger)
		if oerr != nil {
			return nil, fmt.Errorf("openapi validator: %w", oerr)
		}
	}

	_ = time.Second
	return &Dependencies{Config: cfg, Pool: pool, Logger: logger, Validate: v, Handlers: h, Workers: workers, OpenAPI: openAPI}, nil
}
