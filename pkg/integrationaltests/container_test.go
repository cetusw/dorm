package integrationaltests

import (
	"database/sql"
	"database/sql/driver"
	"os"
	"path/filepath"
	"testing"

	"dorm/pkg/infrastructure/di"
)

type MockDriver struct{}

func (d *MockDriver) Open(_ string) (driver.Conn, error) {
	return &MockConn{}, nil
}

type MockConn struct{}

func (c *MockConn) Prepare(_ string) (driver.Stmt, error) { return nil, nil }
func (c *MockConn) Close() error                          { return nil }
func (c *MockConn) Begin() (driver.Tx, error)             { return nil, nil }

func init() {
	sql.Register("test-driver", &MockDriver{})
}

func TestNewContainer(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	configContent := []byte(`{
		"weekStart": "0 9 * * 1"
	}`)
	if err := os.WriteFile(configPath, configContent, 0644); err != nil {
		t.Fatalf("Failed to create temp config: %v", err)
	}

	t.Setenv("DB_DRIVER", "test-driver")
	t.Setenv("DB_USER", "user")
	t.Setenv("DB_PASSWORD", "pass")
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "3306")
	t.Setenv("DB_NAME", "dorm")
	t.Setenv("TZ", "UTC")

	container, err := di.NewContainer(configPath)

	if err != nil {
		t.Fatalf("NewContainer returned error: %v", err)
	}
	defer container.Close()

	if container == nil {
		t.Fatal("Container is nil")
	}

	if container.Config == nil {
		t.Error("Config is nil")
	}

	if container.DB == nil {
		t.Error("DB connection is nil")
	}

	if container.Config.Cron.WeekStart != "0 9 * * 1" {
		t.Errorf("Expected WeekStart '0 9 * * 1', got '%s'", container.Config.Cron.WeekStart)
	}
}
