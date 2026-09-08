package venues_list

import (
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"

	"github.com/zlatovlaska707/avito-kitchen/internal/handler/apierr"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/httpx"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/mapper"
	venueslist "github.com/zlatovlaska707/avito-kitchen/internal/usecase/venues/list"
)

type Handler struct {
	uc        UseCase
	validator *validator.Validate
}

func New(uc UseCase, v *validator.Validate) *Handler {
	return &Handler{uc: uc, validator: v}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	limit, _ := httpx.ParseLimit(r, 20, 100)
	cursorTS, cursorID, _ := httpx.ParseCursor(r)

	var cityID *int
	if v := r.URL.Query().Get("city_id"); v != "" {
		n, _ := strconv.Atoi(v)
		cityID = &n
	}
	var lat, lon *float64
	if v := r.URL.Query().Get("lat"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			lat = &f
		}
	}
	if v := r.URL.Query().Get("lon"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			lon = &f
		}
	}
	var search *string
	if q := r.URL.Query().Get("q"); q != "" {
		search = &q
	}
	sort := r.URL.Query().Get("sort")
	if sort == "" {
		sort = "rating"
	}

	out, err := h.uc.Run(ctx, venueslist.In{
		CityID: cityID, Lat: lat, Lon: lon, Search: search, Sort: sort,
		OnlyOpen: r.URL.Query().Get("only_open") == "true",
		Limit: limit, CursorTS: cursorTS, CursorID: cursorID,
	})
	if err != nil {
		httpx.RespondError(w, ctx, apierr.MapDomainError(err), err)
		return
	}
	items := make([]mapper.Venue, 0, len(out.Items))
	for _, it := range out.Items {
		items = append(items, mapper.VenueFrom(it.Venue, it.IsOpen, it.ETAMinutes, it.DistanceM))
	}
	httpx.RespondJSON(w, http.StatusOK, map[string]any{"items": items, "page_info": mapper.PageFrom(limit, out.HasNext, out.LastTS, out.LastID)})
}
