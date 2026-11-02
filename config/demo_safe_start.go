package config

import (
	"errors"
	"net"
	"strings"
)

// DemoSafeStart is explicit opt-in; ordinary startup remains unchanged.
func demoSafeStart(getenv func(string) string) (bool, error) {
	switch getenv("FOODHUB_DEMO_SAFE_START") {
	case "", "0":
		return false, nil
	case "1":
	default:
		return false, errors.New("FOODHUB_DEMO_SAFE_START must be 0 or 1")
	}
	if getenv("APP_ENV") != "development" || getenv("DB_NAME") != "foodhub_new_structure" || !isLoopbackHost(getenv("DB_HOST")) {
		return false, errors.New("demo safe-start requires development and the local foodhub_new_structure database")
	}
	for _, key := range []string{"GO_ENV", "ENVIRONMENT", "NODE_ENV"} {
		value := getenv(key)
		if value != "" && value != "development" && value != "local" && value != "dev" {
			return false, errors.New("demo safe-start refuses a non-development environment")
		}
	}
	return true, nil
}
func isLoopbackHost(host string) bool {
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(strings.Trim(host, "[]"))
	return ip != nil && ip.IsLoopback()
}
