package server

import (
	"os"
	"testing"

	"github.com/grantbirki/noot/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestCreateDatabaseConfig(t *testing.T) {
	tests := []struct {
		name       string
		setupEnv   func()
		cleanupEnv func()
		expected   storage.Config
	}{
		{
			name: "default configuration",
			setupEnv: func() {
				clearDBEnvVars()
			},
			cleanupEnv: func() {},
			expected: storage.Config{
				Type:     "supabase",
				Database: "",
				Host:     "localhost",
				Port:     5432,
				Username: "",
				Password: "",
				SSLMode:  "prefer",
			},
		},
		{
			name: "postgres configuration",
			setupEnv: func() {
				clearDBEnvVars()
				os.Setenv("DATABASE_PROVIDER", "postgres")
				os.Setenv("DB_HOST", "testhost")
				os.Setenv("DB_PORT", "5433")
				os.Setenv("DB_USER", "testuser")
				os.Setenv("DB_PASS", "testpass")
				os.Setenv("DB_SSLMODE", "require")
			},
			cleanupEnv: func() {
				clearDBEnvVars()
			},
			expected: storage.Config{
				Type:     "postgres",
				Database: "",
				Host:     "testhost",
				Port:     5433,
				Username: "testuser",
				Password: "testpass",
				SSLMode:  "require",
			},
		},
		{
			name: "supabase configuration with URL override",
			setupEnv: func() {
				clearDBEnvVars()
				os.Setenv("DATABASE_PROVIDER", "supabase")
				os.Setenv("SUPABASE_DB_URL", "postgresql://user:pass@host:port/db")
			},
			cleanupEnv: func() {
				clearDBEnvVars()
			},
			expected: storage.Config{
				Type:     "supabase",
				Database: "postgresql://user:pass@host:port/db", // Should be overridden by SUPABASE_DB_URL
				Host:     "localhost",
				Port:     5432,
				Username: "",
				Password: "",
				SSLMode:  "prefer",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment
			tt.setupEnv()
			defer tt.cleanupEnv()

			config := CreateDatabaseConfig()

			assert.Equal(t, tt.expected.Type, config.Type)
			assert.Equal(t, tt.expected.Database, config.Database)
			assert.Equal(t, tt.expected.Host, config.Host)
			assert.Equal(t, tt.expected.Port, config.Port)
			assert.Equal(t, tt.expected.Username, config.Username)
			assert.Equal(t, tt.expected.Password, config.Password)
			assert.Equal(t, tt.expected.SSLMode, config.SSLMode)
		})
	}
}

// Helper function to clear database environment variables
func clearDBEnvVars() {
	os.Unsetenv("DATABASE_PROVIDER")
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("DB_USER")
	os.Unsetenv("DB_PASS")
	os.Unsetenv("DB_SSLMODE")
	os.Unsetenv("SUPABASE_DB_URL")
}
