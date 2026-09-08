package venue_create

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/apierr"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/httpx"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/mapper"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
	venuerepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/venue"
	onboard "github.com/zlatovlaska707/avito-kitchen/internal/usecase/venues/onboard"
)

type Handler struct {
	uc        UseCase
	validator *validator.Validate
}

func New(uc UseCase, v *validator.Validate) *Handler { return &Handler{uc: uc, validator: v} }

type createReq struct {
	Slug             string                  `json:"slug" validate:"required"`
	Name             string                  `json:"name" validate:"required"`
	Description      *string                 `json:"description"`
	CityID           int                     `json:"city_id" validate:"required"`
	AddressText      string                  `json:"address_text" validate:"required"`
	Lat              float64                 `json:"lat" validate:"required"`
	Lon              float64                 `json:"lon" validate:"required"`
	Timezone         string                  `json:"timezone"`
	AvgPrepMinutes   int                     `json:"avg_prep_minutes"`
	DeliveryFeeMinor int64                   `json:"delivery_fee_minor"`
	MinOrderMinor    int64                   `json:"min_order_minor"`
	DeliveryRadiusM  int                     `json:"delivery_radius_m"`
	WorkingHours     []venuerepo.WorkingHour `json:"working_hours"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req createReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.RespondError(w, ctx, apierr.MapDomainError(usecase.ErrValidation), err)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		httpx.RespondError(w, ctx, apierr.MapDomainError(usecase.ErrValidation), err)
		return
	}
	out, err := h.uc.Run(ctx, onboard.In{
		Slug: req.Slug, Name: req.Name, Description: req.Description, CityID: req.CityID,
		AddressText: req.AddressText, Lat: req.Lat, Lon: req.Lon, Timezone: req.Timezone,
		AvgPrepMinutes: req.AvgPrepMinutes, DeliveryFeeMinor: req.DeliveryFeeMinor,
		MinOrderMinor: req.MinOrderMinor, DeliveryRadiusM: req.DeliveryRadiusM, WorkingHours: req.WorkingHours,
	})
	if err != nil {
		httpx.RespondError(w, ctx, apierr.MapDomainError(err), err)
		return
	}
	v := mapper.VenueFrom(out.Venue, false, nil, nil)
	httpx.RespondJSON(w, http.StatusCreated, map[string]any{"venue": v})
}
