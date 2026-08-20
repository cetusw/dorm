package webapp

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestWebAppRoutes(t *testing.T) {
	webRoot := t.TempDir()
	mustWriteFile(t, filepath.Join(webRoot, "index.html"), `<!doctype html><html><head><link rel="stylesheet" href="/app/assets/index-test.css"></head><body>SPA_INDEX</body></html>`)
	mustWriteFile(t, filepath.Join(webRoot, "service-worker.js"), `self.addEventListener("install", () => {});`)
	mustWriteFile(t, filepath.Join(webRoot, "manifest.webmanifest"), `{"name":"dorm"}`)
	mustWriteFile(t, filepath.Join(webRoot, "favicon.svg"), `<svg xmlns="http://www.w3.org/2000/svg"></svg>`)
	mustWriteFile(t, filepath.Join(webRoot, "icons", "app-icon-192.png"), `png`)
	mustWriteFile(t, filepath.Join(webRoot, "assets", "index-test.css"), `body{color:#166534;}`)

	app := fiber.New()
	RegisterRoutes(app, webRoot)

	tests := []struct {
		name             string
		target           string
		wantStatus       int
		wantContentType  string
		wantCacheControl string
		wantBodyContains string
		wantBodyExcludes string
	}{
		{
			name:             "app root serves index html",
			target:           "/app/",
			wantStatus:       http.StatusOK,
			wantContentType:  "text/html",
			wantCacheControl: "no-cache",
			wantBodyContains: "SPA_INDEX",
		},
		{
			name:             "spa deep link serves index html",
			target:           "/app/tasks",
			wantStatus:       http.StatusOK,
			wantContentType:  "text/html",
			wantCacheControl: "no-cache",
			wantBodyContains: "SPA_INDEX",
		},
		{
			name:             "existing asset serves immutable file",
			target:           "/app/assets/index-test.css",
			wantStatus:       http.StatusOK,
			wantContentType:  "text/css",
			wantCacheControl: "public, max-age=31536000, immutable",
			wantBodyContains: "color:#166534",
			wantBodyExcludes: "SPA_INDEX",
		},
		{
			name:             "missing asset returns 404 instead of index",
			target:           "/app/assets/definitely-does-not-exist.css",
			wantStatus:       http.StatusNotFound,
			wantBodyExcludes: "SPA_INDEX",
		},
		{
			name:             "service worker is served without immutable cache",
			target:           "/app/service-worker.js",
			wantStatus:       http.StatusOK,
			wantContentType:  "javascript",
			wantCacheControl: "no-cache",
			wantBodyContains: "self.addEventListener",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.target, nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("app.Test() error = %v", err)
			}
			defer resp.Body.Close()

			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("ReadAll() error = %v", err)
			}

			body := string(bodyBytes)
			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}

			if tt.wantContentType != "" && !strings.Contains(resp.Header.Get("Content-Type"), tt.wantContentType) {
				t.Fatalf("Content-Type = %q, want substring %q", resp.Header.Get("Content-Type"), tt.wantContentType)
			}

			if tt.wantCacheControl != "" && resp.Header.Get("Cache-Control") != tt.wantCacheControl {
				t.Fatalf("Cache-Control = %q, want %q", resp.Header.Get("Cache-Control"), tt.wantCacheControl)
			}

			if tt.wantBodyContains != "" && !strings.Contains(body, tt.wantBodyContains) {
				t.Fatalf("body does not contain %q: %q", tt.wantBodyContains, body)
			}

			if tt.wantBodyExcludes != "" && strings.Contains(body, tt.wantBodyExcludes) {
				t.Fatalf("body unexpectedly contains %q: %q", tt.wantBodyExcludes, body)
			}
		})
	}
}

func mustWriteFile(t *testing.T, filePath string, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", filePath, err)
	}

	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", filePath, err)
	}
}
