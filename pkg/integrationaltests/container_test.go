package integrationaltests

import (
	"database/sql"
	"database/sql/driver"
	"os"
	"path/filepath"
	"testing"

	"dorm/pkg/infrastructure/di"
	"dorm/pkg/infrastructure/spreadsheet"

	"google.golang.org/api/sheets/v4"
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

type FakeSpreadsheetClient struct{}

func (FakeSpreadsheetClient) CreateSheet(string, string) (int64, error) { return 0, nil }
func (FakeSpreadsheetClient) RecreateSheet(string, string) (int64, error) {
	return 0, nil
}
func (FakeSpreadsheetClient) HideSheet(string, int64) error { return nil }
func (FakeSpreadsheetClient) UpdateValues(string, string, [][]interface{}) error {
	return nil
}
func (FakeSpreadsheetClient) BatchUpdateValues(string, []*sheets.ValueRange) error {
	return nil
}
func (FakeSpreadsheetClient) HideSheetsExcept(string, []int64) error { return nil }
func (FakeSpreadsheetClient) BatchUpdate(string, []*sheets.Request) error {
	return nil
}

func TestNewContainer(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	credentialsPath := filepath.Join(tempDir, "google-credentials.json")
	configContent := []byte(`{
		"weekStart": "0 9 * * 1",
		"syncStart": "0 20 * * 0"
	}`)
	if err := os.WriteFile(configPath, configContent, 0644); err != nil {
		t.Fatalf("Failed to create temp config: %v", err)
	}
	if err := os.WriteFile(credentialsPath, []byte(`{}`), 0644); err != nil {
		t.Fatalf("Failed to create temp credentials: %v", err)
	}

	t.Setenv("DB_DRIVER", "test-driver")
	t.Setenv("BOT_TOKEN", "7528543633:AAEIVLeGoadblN9KRr8h6R1L5ZiChzFrACk")
	t.Setenv("GOOGLE_CREDENTIALS", credentialsPath)
	t.Setenv("DB_USER", "user")
	t.Setenv("DB_PASSWORD", "pass")
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "3306")
	t.Setenv("DB_NAME", "dorm")
	t.Setenv("TZ", "UTC")

	container, err := di.NewContainerWithOptions(configPath, di.ContainerOptions{
		NewSpreadsheetClient: func(string) (spreadsheet.Client, error) {
			return FakeSpreadsheetClient{}, nil
		},
	})

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
