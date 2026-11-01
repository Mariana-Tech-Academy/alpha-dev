package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func demoEnvironment(t *testing.T) {
	t.Helper()
	for key, value := range map[string]string{"FOODHUB_DEMO_STATUS": "1", "APP_ENV": "development", "DB_NAME": "foodhub_new_structure", "DB_HOST": "localhost", "GO_ENV": "", "NODE_ENV": "", "ENVIRONMENT": ""} {
		t.Setenv(key, value)
	}
}
func TestLocalDemoGate(t *testing.T) {
	for _, tc := range []struct {
		name, key, value, peer, host, forward string
		want                                  int
	}{
		{name: "loopback enabled", want: 200},
		{name: "off by default", key: "FOODHUB_DEMO_STATUS", want: 404},
		{name: "production", key: "APP_ENV", value: "production", want: 404},
		{name: "production secondary marker", key: "NODE_ENV", value: "production", want: 404},
		{name: "remote database", key: "DB_HOST", value: "db.example.test", want: 404},
		{name: "other database", key: "DB_NAME", value: "shared", want: 404},
		{name: "remote caller", peer: "192.0.2.1:1234", want: 404},
		{name: "remote host", host: "foodhub.example.test", want: 404},
		{name: "forwarded caller", forward: "192.0.2.1", want: 404},
	} {
		t.Run(tc.name, func(t *testing.T) {
			demoEnvironment(t)
			if tc.key != "" {
				t.Setenv(tc.key, tc.value)
			}
			r := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/demo/orders/1/advance", nil)
			r.RemoteAddr = "127.0.0.1:1234"
			if tc.peer != "" {
				r.RemoteAddr = tc.peer
			}
			if tc.host != "" {
				r.Host = tc.host
			}
			if tc.forward != "" {
				r.Header.Set("X-Forwarded-For", tc.forward)
			}
			w := httptest.NewRecorder()
			LocalDemoOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })).ServeHTTP(w, r)
			if w.Code != tc.want {
				t.Fatalf("status %d", w.Code)
			}
		})
	}
}
