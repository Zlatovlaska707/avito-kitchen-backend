package menu_get

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/zlatovlaska707/avito-kitchen/internal/handler/apierr"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/httpx"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/mapper"
	menget "github.com/zlatovlaska707/avito-kitchen/internal/usecase/catalog/menu_get"
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
	out, err := h.uc.Run(ctx, menget.In{VenueRef: chi.URLParam(r, "id")})
	if err != nil {
		httpx.RespondError(w, ctx, apierr.MapDomainError(err), err)
		return
	}
	w.Header().Set("ETag", out.ETag)
	if etagMatches(r.Header.Get("If-None-Match"), out.ETag) {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, mapper.MenuFrom(out.Menu))
}

func etagMatches(ifNoneMatch, etag string) bool {
	if ifNoneMatch == "" || etag == "" {
		return false
	}
	want := normalizeETag(etag)
	for _, part := range strings.Split(ifNoneMatch, ",") {
		if normalizeETag(part) == want {
			return true
		}
	}
	return false
}

func normalizeETag(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "W/")
	return strings.Trim(v, `"`)
}
