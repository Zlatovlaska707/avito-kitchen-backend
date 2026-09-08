package system

import (
	"net/http"
	"os"
)

func resolveOpenAPISpec() string {
	for _, p := range []string{"openapi.yaml", "api/openapi.yaml", "./openapi.yaml"} {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return "openapi.yaml"
}

func SwaggerUI() http.HandlerFunc {
	return SwaggerUIWithSpec(resolveOpenAPISpec())
}

func SwaggerUIWithSpec(specPath string) http.HandlerFunc {
	spec, err := os.ReadFile(specPath) //nolint:gosec // spec path is chosen from a fixed allowlist
	if err != nil {
		return func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "openapi spec not found", http.StatusNotFound)
		}
	}
	specJSON := spec

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/swagger/openapi.yaml", "/swagger/openapi.json":
			w.Header().Set("Content-Type", "application/yaml")
			_, _ = w.Write(specJSON)
			return
		case "/swagger", "/swagger/":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(swaggerHTML))
		default:
			http.NotFound(w, r)
		}
	}
}

const swaggerHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8"/>
  <title>Avito Kitchen API</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css"/>
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
<script>
  window.onload = () => {
    SwaggerUIBundle({
      url: '/swagger/openapi.yaml',
      dom_id: '#swagger-ui',
      deepLinking: true,
      presets: [SwaggerUIBundle.presets.apis],
    });
  };
</script>
</body>
</html>`
