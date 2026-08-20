package webapp

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
)

const (
	indexCacheControl          = "no-cache"
	serviceWorkerCacheControl  = "no-cache"
	immutableAssetCacheControl = "public, max-age=31536000, immutable"
)

var rootStaticFiles = map[string]string{
	"/app/favicon.svg":          "favicon.svg",
	"/app/icons.svg":            "icons.svg",
	"/app/manifest.webmanifest": "manifest.webmanifest",
	"/app/service-worker.js":    "service-worker.js",
}

func RegisterRoutes(app *fiber.App, webRoot string) {
	indexPath := filepath.Join(webRoot, "index.html")
	assetsDir := filepath.Join(webRoot, "assets")
	iconsDir := filepath.Join(webRoot, "icons")

	registerFileRoute(app, "/app", indexPath, indexCacheControl)
	registerFileRoute(app, "/app/", indexPath, indexCacheControl)
	registerFileRoute(app, "/app/service-worker.js", filepath.Join(webRoot, "service-worker.js"), serviceWorkerCacheControl)
	registerDirectoryRoute(app, "/app/assets/*", assetsDir, immutableAssetCacheControl)
	registerDirectoryRoute(app, "/app/icons/*", iconsDir, indexCacheControl)

	for route, relativePath := range rootStaticFiles {
		if route == "/app/service-worker.js" {
			continue
		}

		registerFileRoute(app, route, filepath.Join(webRoot, relativePath), indexCacheControl)
	}

	registerFileRoute(app, "/app/*", indexPath, indexCacheControl)
}

func registerFileRoute(app *fiber.App, route string, filePath string, cacheControl string) {
	handler := func(c *fiber.Ctx) error {
		return sendExistingFile(c, filePath, cacheControl)
	}

	app.Get(route, handler)
	app.Head(route, handler)
}

func registerDirectoryRoute(app *fiber.App, route string, rootDir string, cacheControl string) {
	handler := func(c *fiber.Ctx) error {
		targetPath, err := resolvePathWithinRoot(rootDir, c.Params("*"))
		if err != nil {
			return fiber.ErrNotFound
		}

		return sendExistingFile(c, targetPath, cacheControl)
	}

	app.Get(route, handler)
	app.Head(route, handler)
}

func sendExistingFile(c *fiber.Ctx, filePath string, cacheControl string) error {
	info, err := os.Stat(filePath)
	if err != nil || info.IsDir() {
		return fiber.ErrNotFound
	}

	c.Set(fiber.HeaderCacheControl, cacheControl)
	return c.SendFile(filePath)
}

func resolvePathWithinRoot(rootDir string, requestPath string) (string, error) {
	cleanRequestPath := strings.TrimPrefix(filepath.Clean("/"+requestPath), "/")
	targetPath := filepath.Join(rootDir, filepath.FromSlash(cleanRequestPath))

	relativePath, err := filepath.Rel(rootDir, targetPath)
	if err != nil {
		return "", fmt.Errorf("resolve static file path: %w", err)
	}

	if relativePath == ".." || strings.HasPrefix(relativePath, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes static root")
	}

	return targetPath, nil
}
