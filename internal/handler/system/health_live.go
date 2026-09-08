package system

import (
	"net/http"

	"github.com/zlatovlaska707/avito-kitchen/internal/handler/httpx"
)

func Live(w http.ResponseWriter, _ *http.Request) {
	httpx.RespondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
