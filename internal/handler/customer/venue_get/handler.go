package venue_get

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/zlatovlaska707/avito-kitchen/internal/handler/apierr"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/httpx"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/mapper"
	venueget "github.com/zlatovlaska707/avito-kitchen/internal/usecase/venues/get"
)

type Handler struct {
	uc        UseCase
	validator *validator.Validate
}

func New(uc UseCase, v *validator.Validate) *Handler {
	return &Handler{uc: uc, validator: v}
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	out, err := h.uc.Run(ctx, venueget.In{VenueRef: chi.URLParam(r, "id")})
	if err != nil {
		httpx.RespondError(w, ctx, apierr.MapDomainError(err), err)
		return
	}
	v := mapper.VenueFrom(out.Venue, out.IsOpen, &out.ETAMinutes, nil)
	httpx.RespondJSON(w, http.StatusOK, mapper.VenueDetails{
		Venue: v, Description: out.Venue.Description, Status: out.Venue.Status,
		AcceptsOrders: out.Venue.AcceptsOrders, CityID: out.Venue.CityID, AddressText: out.Venue.AddressText,
		Lat: out.Venue.Lat, Lon: out.Venue.Lon, Timezone: out.Venue.Timezone,
		WorkingHours: mapper.WorkingHoursFrom(out.Hours), ClosedReason: out.ClosedReason,
	})
}
