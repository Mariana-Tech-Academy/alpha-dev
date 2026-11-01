package middleware

import (
	"alpha-dev/utils"
	"net"
	"net/http"
	"os"
	"strings"
)

func loopbackHost(host string) bool {
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(strings.Trim(host, "[]"))
	return ip != nil && ip.IsLoopback()
}

// Fail closed: the demo route must never target shared or production data.
func LocalDemoEnabled() bool {
	if os.Getenv("FOODHUB_DEMO_STATUS") != "1" || os.Getenv("APP_ENV") != "development" || os.Getenv("DB_NAME") != "foodhub_new_structure" || !loopbackHost(os.Getenv("DB_HOST")) {
		return false
	}
	for _, key := range []string{"GO_ENV", "ENVIRONMENT", "NODE_ENV"} {
		value := os.Getenv(key)
		if value != "" && value != "development" && value != "local" && value != "dev" {
			return false
		}
	}
	return true
}

func LocalDemoOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		peer, _, err := net.SplitHostPort(r.RemoteAddr)
		host := r.Host
		if parsed, _, e := net.SplitHostPort(host); e == nil {
			host = parsed
		}
		if !LocalDemoEnabled() || err != nil || !loopbackHost(peer) || !loopbackHost(host) || r.Header.Get("Forwarded") != "" || r.Header.Get("X-Forwarded-For") != "" || r.Header.Get("X-Forwarded-Host") != "" {
			utils.Error(w, http.StatusNotFound, "route not found")
			return
		}
		next.ServeHTTP(w, r)
	})
}
