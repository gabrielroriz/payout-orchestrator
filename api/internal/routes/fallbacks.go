package routes

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/gabrielroriz/payout-orchestrator/api/internal/httpjson"
)

// methods are the HTTP methods checked when listing what a path allows.
var methods = []string{
	http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch,
	http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace,
}

// notFound answers a request whose path matches no route.
func notFound(w http.ResponseWriter, r *http.Request) {
	httpjson.Error(w, http.StatusNotFound, httpjson.CodeNotFound, "not found")
}

// methodNotAllowed answers a request whose path exists but not with its method.
func methodNotAllowed(routes chi.Routes) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// chi sets Allow only in its default 405 handler, so list the path's methods ourselves.
		var allowed []string
		for _, m := range methods {
			if routes.Match(chi.NewRouteContext(), m, r.URL.Path) {
				allowed = append(allowed, m)
			}
		}

		// Tell the client which methods it can use instead, as HTTP requires on a 405.
		w.Header().Set("Allow", strings.Join(allowed, ", "))
		httpjson.Error(w, http.StatusMethodNotAllowed, httpjson.CodeMethodNotAllowed, "method not allowed")
	}
}
