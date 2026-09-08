package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/zlatovlaska707/avito-kitchen/internal/handler/middleware"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/system"
	"github.com/zlatovlaska707/avito-kitchen/internal/metrics"
)

func registerRoutes(r chi.Router, deps *Dependencies) {
	common := chi.Chain(
		middleware.RequestID,
		middleware.Logger(deps.Logger),
		middleware.Recover(deps.Logger),
		middleware.WithHandlerTimeout(deps.Config.HTTP.HandlerTimeout),
	)

	openAPI := func() func(http.Handler) http.Handler {
		if deps.OpenAPI == nil {
			return func(next http.Handler) http.Handler { return next }
		}
		return deps.OpenAPI.Middleware
	}()

	r.Group(func(r chi.Router) {
		r.Use(common...)
		r.With(openAPI).Get("/health/live", system.Live)
		r.With(openAPI).Get("/health/ready", system.Ready(deps.Pool))
	})

	r.Handle("/metrics", metrics.Handler())

	if deps.Config.Swagger.Enabled {
		swagger := system.SwaggerUI()
		r.Get("/swagger", swagger)
		r.Get("/swagger/", swagger)
		r.Get("/swagger/openapi.yaml", swagger)
		r.Get("/swagger/openapi.json", swagger)
	}

	h := deps.Handlers

	r.Group(func(r chi.Router) {
		r.Use(common...)
		r.With(middleware.BodyLimit(deps.Config.HTTP.MaxBodyPublic)).
			With(middleware.NewRateLimiter(deps.Config.Rate.RPSPublic, deps.Config.Rate.RPSPublic, middleware.IPKey).Middleware()).
			Route("/api/v1", func(r chi.Router) {
				r.Use(middleware.CORS(nil))
				r.With(openAPI).Get("/venues", h.CustomerVenuesList.List)
				r.With(openAPI).Get("/venues/{id}", h.CustomerVenueGet.Get)
				r.With(openAPI).Get("/venues/{id}/menu", h.CustomerMenuGet.Get)

				r.Group(func(r chi.Router) {
					r.Use(middleware.CustomerIdentity(), openAPI)
					r.Get("/cart", h.CustomerCartGet.Get)
					r.Post("/cart/items", h.CustomerCartItemAdd.Add)
					r.Patch("/cart/items/{id}", h.CustomerCartItemUpdate.Update)
					r.Delete("/cart", h.CustomerCartClear.Clear)
					r.Post("/orders", h.CustomerOrderCreate.Create)
					r.Get("/orders", h.CustomerOrderList.List)
					r.Get("/orders/{id}", h.CustomerOrderGet.Get)
					r.Post("/orders/{id}/cancel", h.CustomerOrderCancel.Cancel)
				})
			})
	})

	r.Group(func(r chi.Router) {
		r.Use(common...)
		r.With(middleware.BodyLimit(deps.Config.HTTP.MaxBodyBytes)).
			With(middleware.NewRateLimiter(deps.Config.Rate.RPSPartner, deps.Config.Rate.RPSPartner, middleware.IPKey).Middleware()).
			Route("/partner/v1", func(r chi.Router) {
				r.Use(middleware.PartnerAuth("", deps.Handlers.APIKeyRepo), openAPI)
				r.With(middleware.RequirePartnerScope(middleware.ScopeMenuWrite)).Put("/menu", h.PartnerMenuSync.Sync)
				r.With(middleware.RequirePartnerScope(middleware.ScopeMenuWrite)).Patch("/menu/items/{externalId}/availability", h.PartnerItemAvailability.Set)
				r.With(middleware.RequirePartnerScope(middleware.ScopeMenuWrite)).Patch("/venue/status", h.PartnerVenueStatusSet.Set)
				r.With(middleware.RequirePartnerScope(middleware.ScopeOrdersWrite)).Get("/orders", h.PartnerOrderList.List)
				r.With(middleware.RequirePartnerScope(middleware.ScopeOrdersWrite)).Get("/orders/{id}", h.PartnerOrderGet.Get)
				r.With(middleware.RequirePartnerScope(middleware.ScopeOrdersWrite)).Post("/orders/{id}/accept", h.PartnerOrderAccept.Accept)
				r.With(middleware.RequirePartnerScope(middleware.ScopeOrdersWrite)).Post("/orders/{id}/reject", h.PartnerOrderReject.Reject)
				r.With(middleware.RequirePartnerScope(middleware.ScopeOrdersWrite)).Post("/orders/{id}/ready", h.PartnerOrderReady.Ready)
				r.With(middleware.RequirePartnerScope(middleware.ScopeOrdersWrite)).Post("/orders/{id}/handover", h.PartnerOrderHandover.Handover)
				r.With(middleware.RequirePartnerScope(middleware.ScopeOrdersWrite)).Post("/webhooks", h.PartnerWebhookCreate.Create)
				r.With(middleware.RequirePartnerScope(middleware.ScopeOrdersWrite)).Delete("/webhooks/{id}", h.PartnerWebhookDelete.Delete)
			})
	})

	r.Group(func(r chi.Router) {
		r.Use(common...)
		r.With(middleware.BodyLimit(deps.Config.HTTP.MaxBodyBytes)).
			With(middleware.NewRateLimiter(deps.Config.Rate.RPSPartner, deps.Config.Rate.RPSPartner, middleware.IPKey).Middleware()).
			Route("/admin/v1", func(r chi.Router) {
				r.Use(middleware.AdminAuth(deps.Config.AdminAPIToken), openAPI)
				r.Post("/venues", h.AdminVenueCreate.Create)
				r.Patch("/venues/{id}", h.AdminVenueUpdate.Update)
				r.Post("/venues/{id}/api-keys", h.AdminAPIKeyIssue.Issue)
				r.Delete("/venues/{venueId}/api-keys/{keyId}", h.AdminAPIKeyRevoke.Revoke)
				r.Get("/orders", h.AdminOrderList.List)
			})
	})

	_ = http.StatusOK
}
