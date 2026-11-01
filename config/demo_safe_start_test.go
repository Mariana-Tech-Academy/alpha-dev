package config

import "testing"

func TestDemoSafeStartConfiguration(t *testing.T) {
	for _, tc := range []struct {
		name, key, value string
		enabled, invalid bool
	}{
		{name: "disabled by default"},
		{name: "explicit zero", key: "FOODHUB_DEMO_SAFE_START", value: "0"},
		{name: "enabled locally", enabled: true},
		{name: "IPv4", key: "DB_HOST", value: "127.0.0.1", enabled: true},
		{name: "IPv6", key: "DB_HOST", value: "::1", enabled: true},
		{name: "production", key: "APP_ENV", value: "production", invalid: true},
		{name: "missing environment", key: "APP_ENV", invalid: true},
		{name: "conflicting environment", key: "GO_ENV", value: "production", invalid: true},
		{name: "conflicting node environment", key: "NODE_ENV", value: "production", invalid: true},
		{name: "conflicting generic environment", key: "ENVIRONMENT", value: "staging", invalid: true},
		{name: "wrong database", key: "DB_NAME", value: "other", invalid: true},
		{name: "remote database", key: "DB_HOST", value: "db.example.test", invalid: true},
		{name: "lookalike local hostname", key: "DB_HOST", value: "localhost.example.test", invalid: true},
		{name: "missing database host", key: "DB_HOST", invalid: true},
		{name: "invalid flag", key: "FOODHUB_DEMO_SAFE_START", value: "true", invalid: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			env := map[string]string{"FOODHUB_DEMO_SAFE_START": "1", "APP_ENV": "development", "DB_NAME": "foodhub_new_structure", "DB_HOST": "localhost"}
			if tc.name == "disabled by default" {
				delete(env, "FOODHUB_DEMO_SAFE_START")
				env["APP_ENV"] = "production"
				env["DB_HOST"] = "remote"
			}
			if tc.key != "" {
				env[tc.key] = tc.value
			}
			enabled, err := demoSafeStart(func(key string) string { return env[key] })
			if enabled != tc.enabled || (err != nil) != tc.invalid {
				t.Fatalf("enabled=%v err=%v", enabled, err)
			}
		})
	}
}
