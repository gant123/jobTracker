package middleware

import (
	"net/http"
	"strings"

	"github.com/gant123/jobTracker/internal/config"
)

func CORS(cfg *config.Config) func(http.Handler) http.Handler {
	allowed := splitAndTrim(cfg.AllowedOrigins)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Always vary on Origin for proxies/CDNs
			w.Header().Add("Vary", "Origin")
			w.Header().Add("Vary", "Access-Control-Request-Method")
			w.Header().Add("Vary", "Access-Control-Request-Headers")

			// If no Origin, it's not a CORS request
			if origin != "" && (originAllowed(origin, allowed) || originAllowedWildcard(allowed)) {
				// With credentials, you must echo the exact Origin (not *)
				if originAllowedWildcard(allowed) {
					// Only use wildcard if you are NOT using credentials.
					// Since we allow credentials, prefer echoing the origin when present.
					w.Header().Set("Access-Control-Allow-Origin", origin)
				} else {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}

				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

				// Reflect requested headers or set a conservative allow list
				reqHeaders := r.Header.Get("Access-Control-Request-Headers")
				if reqHeaders != "" {
					w.Header().Set("Access-Control-Allow-Headers", reqHeaders)
				} else {
					w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")
				}

				// Optional: expose headers the client may need to read
				// w.Header().Set("Access-Control-Expose-Headers", "Authorization")
			}

			// Short-circuit preflight
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func splitAndTrim(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func originAllowed(origin string, allowed []string) bool {
	for _, a := range allowed {
		if origin == a {
			return true
		}
	}
	return false
}

func originAllowedWildcard(allowed []string) bool {
	for _, a := range allowed {
		if a == "*" {
			return true
		}
	}
	return false
}
