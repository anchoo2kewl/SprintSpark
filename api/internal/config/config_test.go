package config

import (
	"testing"
	"time"
)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		setEnv       bool
		want         string
	}{
		{
			name:         "returns default when env not set",
			key:          "TEST_GET_ENV_MISSING",
			defaultValue: "default_val",
			setEnv:       false,
			want:         "default_val",
		},
		{
			name:         "returns env value when set",
			key:          "TEST_GET_ENV_SET",
			defaultValue: "default_val",
			envValue:     "custom_val",
			setEnv:       true,
			want:         "custom_val",
		},
		{
			name:         "returns default when env is empty string",
			key:          "TEST_GET_ENV_EMPTY",
			defaultValue: "fallback",
			envValue:     "",
			setEnv:       true,
			want:         "fallback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				t.Setenv(tt.key, tt.envValue)
			}

			got := getEnv(tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("getEnv(%q, %q) = %q, want %q", tt.key, tt.defaultValue, got, tt.want)
			}
		})
	}
}

func TestGetEnvAsInt(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue int
		envValue     string
		setEnv       bool
		want         int
	}{
		{
			name:         "returns default when env not set",
			key:          "TEST_INT_MISSING",
			defaultValue: 42,
			setEnv:       false,
			want:         42,
		},
		{
			name:         "returns parsed int when valid",
			key:          "TEST_INT_VALID",
			defaultValue: 10,
			envValue:     "100",
			setEnv:       true,
			want:         100,
		},
		{
			name:         "returns default when env is empty",
			key:          "TEST_INT_EMPTY",
			defaultValue: 5,
			envValue:     "",
			setEnv:       true,
			want:         5,
		},
		{
			name:         "returns default when env is not a number",
			key:          "TEST_INT_INVALID",
			defaultValue: 7,
			envValue:     "not-a-number",
			setEnv:       true,
			want:         7,
		},
		{
			name:         "handles zero value",
			key:          "TEST_INT_ZERO",
			defaultValue: 99,
			envValue:     "0",
			setEnv:       true,
			want:         0,
		},
		{
			name:         "handles negative value",
			key:          "TEST_INT_NEGATIVE",
			defaultValue: 10,
			envValue:     "-5",
			setEnv:       true,
			want:         -5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				t.Setenv(tt.key, tt.envValue)
			}

			got := getEnvAsInt(tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("getEnvAsInt(%q, %d) = %d, want %d", tt.key, tt.defaultValue, got, tt.want)
			}
		})
	}
}

func TestGetEnvAsSlice(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue []string
		envValue     string
		setEnv       bool
		want         []string
	}{
		{
			name:         "returns default when env not set",
			key:          "TEST_SLICE_MISSING",
			defaultValue: []string{"a", "b"},
			setEnv:       false,
			want:         []string{"a", "b"},
		},
		{
			name:         "parses comma-separated values",
			key:          "TEST_SLICE_CSV",
			defaultValue: []string{"default"},
			envValue:     "http://localhost:3000,http://localhost:5173",
			setEnv:       true,
			want:         []string{"http://localhost:3000", "http://localhost:5173"},
		},
		{
			name:         "trims whitespace around values",
			key:          "TEST_SLICE_WHITESPACE",
			defaultValue: []string{"default"},
			envValue:     " foo , bar , baz ",
			setEnv:       true,
			want:         []string{"foo", "bar", "baz"},
		},
		{
			name:         "single value without comma",
			key:          "TEST_SLICE_SINGLE",
			defaultValue: []string{"default"},
			envValue:     "http://example.com",
			setEnv:       true,
			want:         []string{"http://example.com"},
		},
		{
			name:         "returns default when env is empty",
			key:          "TEST_SLICE_EMPTY",
			defaultValue: []string{"fallback"},
			envValue:     "",
			setEnv:       true,
			want:         []string{"fallback"},
		},
		{
			name:         "returns default when only commas and whitespace",
			key:          "TEST_SLICE_ONLY_COMMAS",
			defaultValue: []string{"default"},
			envValue:     ", , ,",
			setEnv:       true,
			want:         []string{"default"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				t.Setenv(tt.key, tt.envValue)
			}

			got := getEnvAsSlice(tt.key, tt.defaultValue)

			if len(got) != len(tt.want) {
				t.Fatalf("getEnvAsSlice(%q) returned %d elements, want %d: got=%v, want=%v",
					tt.key, len(got), len(tt.want), got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("getEnvAsSlice(%q)[%d] = %q, want %q", tt.key, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestSplitCommaSeparated(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "simple split",
			input: "a,b,c",
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "single element",
			input: "only",
			want:  []string{"only"},
		},
		{
			name:  "empty string",
			input: "",
			want:  nil,
		},
		{
			name:  "trailing comma produces empty trailing part",
			input: "a,b,",
			want:  []string{"a", "b"},
		},
		{
			name:  "leading comma produces empty leading part",
			input: ",a,b",
			want:  []string{"", "a", "b"},
		},
		{
			name:  "consecutive commas",
			input: "a,,b",
			want:  []string{"a", "", "b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitCommaSeparated(tt.input)

			if len(got) != len(tt.want) {
				t.Fatalf("splitCommaSeparated(%q) returned %d elements, want %d: got=%v, want=%v",
					tt.input, len(got), len(tt.want), got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("splitCommaSeparated(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestTrimSpace(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no whitespace",
			input: "hello",
			want:  "hello",
		},
		{
			name:  "leading spaces",
			input: "   hello",
			want:  "hello",
		},
		{
			name:  "trailing spaces",
			input: "hello   ",
			want:  "hello",
		},
		{
			name:  "both sides",
			input: "  hello  ",
			want:  "hello",
		},
		{
			name:  "tabs and newlines",
			input: "\t\n hello \r\n\t",
			want:  "hello",
		},
		{
			name:  "all whitespace",
			input: "   \t\n  ",
			want:  "",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "internal whitespace preserved",
			input: "  hello world  ",
			want:  "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := trimSpace(tt.input)
			if got != tt.want {
				t.Errorf("trimSpace(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestJWTExpiry(t *testing.T) {
	tests := []struct {
		name       string
		expiryHrs  int
		wantExpiry time.Duration
	}{
		{
			name:       "default 24 hours",
			expiryHrs:  24,
			wantExpiry: 24 * time.Hour,
		},
		{
			name:       "1 hour",
			expiryHrs:  1,
			wantExpiry: 1 * time.Hour,
		},
		{
			name:       "72 hours",
			expiryHrs:  72,
			wantExpiry: 72 * time.Hour,
		},
		{
			name:       "zero hours",
			expiryHrs:  0,
			wantExpiry: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{JWTExpiryHours: tt.expiryHrs}
			got := cfg.JWTExpiry()
			if got != tt.wantExpiry {
				t.Errorf("JWTExpiry() = %v, want %v", got, tt.wantExpiry)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	t.Run("default values in development", func(t *testing.T) {
		// Ensure ENV is not set to production (to avoid Fatal)
		t.Setenv("ENV", "development")

		cfg := Load()

		if cfg.Port != "8080" {
			t.Errorf("Default Port = %q, want %q", cfg.Port, "8080")
		}
		if cfg.Env != "development" {
			t.Errorf("Default Env = %q, want %q", cfg.Env, "development")
		}
		if cfg.DBPath != "./data/sprintspark.db" {
			t.Errorf("Default DBPath = %q, want %q", cfg.DBPath, "./data/sprintspark.db")
		}
		if cfg.MigrationsPath != "./internal/db/migrations" {
			t.Errorf("Default MigrationsPath = %q, want %q", cfg.MigrationsPath, "./internal/db/migrations")
		}
		if cfg.JWTExpiryHours != 24 {
			t.Errorf("Default JWTExpiryHours = %d, want 24", cfg.JWTExpiryHours)
		}
		if cfg.RateLimitRequests != 100 {
			t.Errorf("Default RateLimitRequests = %d, want 100", cfg.RateLimitRequests)
		}
		if cfg.RateLimitWindowMinutes != 15 {
			t.Errorf("Default RateLimitWindowMinutes = %d, want 15", cfg.RateLimitWindowMinutes)
		}
		if cfg.LogLevel != "info" {
			t.Errorf("Default LogLevel = %q, want %q", cfg.LogLevel, "info")
		}
		if !cfg.EnableSQLLog {
			t.Error("EnableSQLLog should be true in development")
		}
		if cfg.DBQueryTimeout != 5*time.Second {
			t.Errorf("Default DBQueryTimeout = %v, want %v", cfg.DBQueryTimeout, 5*time.Second)
		}
		if len(cfg.CORSAllowedOrigins) != 2 {
			t.Errorf("Default CORSAllowedOrigins length = %d, want 2", len(cfg.CORSAllowedOrigins))
		}
	})

	t.Run("custom values via env vars", func(t *testing.T) {
		t.Setenv("PORT", "9090")
		t.Setenv("ENV", "staging")
		t.Setenv("DB_PATH", "/tmp/test.db")
		t.Setenv("MIGRATIONS_PATH", "/migrations")
		t.Setenv("JWT_SECRET", "my-super-secret")
		t.Setenv("JWT_EXPIRY_HOURS", "48")
		t.Setenv("CORS_ALLOWED_ORIGINS", "https://example.com,https://api.example.com")
		t.Setenv("RATE_LIMIT_REQUESTS", "50")
		t.Setenv("RATE_LIMIT_WINDOW_MINUTES", "10")
		t.Setenv("LOG_LEVEL", "debug")
		t.Setenv("DB_QUERY_TIMEOUT_SECONDS", "10")

		cfg := Load()

		if cfg.Port != "9090" {
			t.Errorf("Port = %q, want %q", cfg.Port, "9090")
		}
		if cfg.Env != "staging" {
			t.Errorf("Env = %q, want %q", cfg.Env, "staging")
		}
		if cfg.DBPath != "/tmp/test.db" {
			t.Errorf("DBPath = %q, want %q", cfg.DBPath, "/tmp/test.db")
		}
		if cfg.MigrationsPath != "/migrations" {
			t.Errorf("MigrationsPath = %q, want %q", cfg.MigrationsPath, "/migrations")
		}
		if cfg.JWTSecret != "my-super-secret" {
			t.Errorf("JWTSecret = %q, want %q", cfg.JWTSecret, "my-super-secret")
		}
		if cfg.JWTExpiryHours != 48 {
			t.Errorf("JWTExpiryHours = %d, want 48", cfg.JWTExpiryHours)
		}
		if cfg.RateLimitRequests != 50 {
			t.Errorf("RateLimitRequests = %d, want 50", cfg.RateLimitRequests)
		}
		if cfg.RateLimitWindowMinutes != 10 {
			t.Errorf("RateLimitWindowMinutes = %d, want 10", cfg.RateLimitWindowMinutes)
		}
		if cfg.LogLevel != "debug" {
			t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, "debug")
		}
		if cfg.DBQueryTimeout != 10*time.Second {
			t.Errorf("DBQueryTimeout = %v, want %v", cfg.DBQueryTimeout, 10*time.Second)
		}
		if len(cfg.CORSAllowedOrigins) != 2 {
			t.Fatalf("CORSAllowedOrigins length = %d, want 2", len(cfg.CORSAllowedOrigins))
		}
		if cfg.CORSAllowedOrigins[0] != "https://example.com" {
			t.Errorf("CORSAllowedOrigins[0] = %q, want %q", cfg.CORSAllowedOrigins[0], "https://example.com")
		}
		if cfg.CORSAllowedOrigins[1] != "https://api.example.com" {
			t.Errorf("CORSAllowedOrigins[1] = %q, want %q", cfg.CORSAllowedOrigins[1], "https://api.example.com")
		}
	})

	t.Run("EnableSQLLog false in non-development", func(t *testing.T) {
		t.Setenv("ENV", "staging")
		t.Setenv("ENABLE_SQL_LOG", "false")

		cfg := Load()

		if cfg.EnableSQLLog {
			t.Error("EnableSQLLog should be false in non-development without ENABLE_SQL_LOG=true")
		}
	})

	t.Run("EnableSQLLog override in non-development", func(t *testing.T) {
		t.Setenv("ENV", "staging")
		t.Setenv("ENABLE_SQL_LOG", "true")

		cfg := Load()

		if !cfg.EnableSQLLog {
			t.Error("EnableSQLLog should be true when ENABLE_SQL_LOG=true even in staging")
		}
	})
}

func TestInitLogger(t *testing.T) {
	tests := []struct {
		name     string
		env      string
		logLevel string
		wantErr  bool
	}{
		{
			name:     "development logger",
			env:      "development",
			logLevel: "debug",
			wantErr:  false,
		},
		{
			name:     "production logger",
			env:      "production",
			logLevel: "info",
			wantErr:  false,
		},
		{
			name:     "production with warn level",
			env:      "production",
			logLevel: "warn",
			wantErr:  false,
		},
		{
			name:     "development with empty log level",
			env:      "development",
			logLevel: "",
			wantErr:  false,
		},
		{
			name:     "invalid log level falls back to default",
			env:      "development",
			logLevel: "not-a-level",
			wantErr:  false,
		},
		{
			name:     "production with error level",
			env:      "production",
			logLevel: "error",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := InitLogger(tt.env, tt.logLevel)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("InitLogger(%q, %q) returned unexpected error: %v", tt.env, tt.logLevel, err)
			}

			if logger == nil {
				t.Fatal("Expected non-nil logger")
			}

			// Verify logger can write without panicking
			logger.Info("test message")
			logger.Sync()
		})
	}
}

func TestMustInitLogger(t *testing.T) {
	t.Run("succeeds with valid config", func(t *testing.T) {
		logger := MustInitLogger("development", "info")
		if logger == nil {
			t.Fatal("Expected non-nil logger from MustInitLogger")
		}
		logger.Info("test from MustInitLogger")
		logger.Sync()
	})
}
