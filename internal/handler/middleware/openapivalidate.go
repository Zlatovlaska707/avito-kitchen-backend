package middleware

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/legacy"

	"github.com/zlatovlaska707/avito-kitchen/internal/handler/httpx"
	openapispec "github.com/zlatovlaska707/avito-kitchen/internal/openapi"
)

// OpenAPIValidator validates HTTP requests and/or responses against the OpenAPI spec.
type OpenAPIValidator struct {
	router           routers.Router
	validateRequest  bool
	validateResponse bool
	logger           *slog.Logger
}

func NewOpenAPIValidator(validateRequest, validateResponse bool, logger *slog.Logger) (*OpenAPIValidator, error) {
	doc, err := openapispec.Load()
	if err != nil {
		return nil, err
	}
	router, err := legacy.NewRouter(doc)
	if err != nil {
		return nil, err
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &OpenAPIValidator{
		router:           router,
		validateRequest:  validateRequest,
		validateResponse: validateResponse,
		logger:           logger,
	}, nil
}

func (v *OpenAPIValidator) Middleware(next http.Handler) http.Handler {
	if v == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route, pathParams, err := v.router.FindRoute(r)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		opts := &openapi3filter.Options{
			AuthenticationFunc: openapi3filter.NoopAuthenticationFunc,
		}
		input := &openapi3filter.RequestValidationInput{
			Request:    r,
			PathParams: pathParams,
			Route:      route,
			Options:    opts,
		}

		if v.validateRequest {
			var bodyBytes []byte
			if r.Body != nil {
				bodyBytes, err = io.ReadAll(r.Body)
				if err != nil {
					httpx.RespondError(w, r.Context(), httpx.APIError{
						Status:  http.StatusBadRequest,
						Code:    "validation_failed",
						Message: "Validation failed",
					}, err)
					return
				}
				r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			}
			if err := openapi3filter.ValidateRequest(r.Context(), input); err != nil {
				v.logger.Warn("openapi request validation failed", "path", r.URL.Path, "error", err)
				httpx.RespondError(w, r.Context(), httpx.APIError{
					Status:  http.StatusBadRequest,
					Code:    "validation_failed",
					Message: "Validation failed",
				}, err)
				return
			}
			if bodyBytes != nil {
				r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			}
		}

		if !v.validateResponse {
			next.ServeHTTP(w, r)
			return
		}

		rec := &responseCapture{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		respInput := &openapi3filter.ResponseValidationInput{
			RequestValidationInput: input,
			Status:                 rec.status,
			Header:                 rec.Header(),
			Body:                   io.NopCloser(bytes.NewReader(rec.body.Bytes())),
		}
		if err := openapi3filter.ValidateResponse(r.Context(), respInput); err != nil {
			v.logger.Error("openapi response validation failed", "path", r.URL.Path, "status", rec.status, "error", err)
		}
	})
}

type responseCapture struct {
	http.ResponseWriter
	status int
	body   bytes.Buffer
}

func (r *responseCapture) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseCapture) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	_, _ = r.body.Write(b)
	return r.ResponseWriter.Write(b)
}
