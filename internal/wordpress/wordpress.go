package wordpress

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// WordPress manages WordPress installations via WP-CLI
type WordPress struct {
	Path string
}

// New creates a WordPress manager for a site
func New(path string) *WordPress {
	return &WordPress{Path: path}
}

// Install downloads and installs WordPress
func (wp *WordPress) Install(url, title, adminUser, adminEmail, adminPass string) error {
	// Download WordPress
	if err := wp.run("core", "download"); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	// Install WordPress
	args := []string{
		"core", "install",
		"--url=" + url,
		"--title=" + title,
		"--admin_user=" + adminUser,
		"--admin_email=" + adminEmail,
		"--admin_password=" + adminPass,
		"--skip-email",
	}
	if err := wp.run(args...); err != nil {
		return fmt.Errorf("install failed: %w", err)
	}

	return nil
}

// CreateConfig creates wp-config.php
func (wp *WordPress) CreateConfig(dbName, dbUser, dbPass, dbHost string) error {
	args := []string{
		"config", "create",
		"--dbname=" + dbName,
		"--dbuser=" + dbUser,
		"--dbpass=" + dbPass,
		"--dbhost=" + dbHost,
	}
	return wp.run(args...)
}

// AutoTune applies performance optimizations
func (wp *WordPress) AutoTune() error {
	configs := map[string]string{
		"WP_MEMORY_LIMIT":     "'256M'",
		"WP_MAX_MEMORY_LIMIT": "'512M'",
		"WP_POST_REVISIONS":   "5",
		"AUTOSAVE_INTERVAL":   "120",
		"EMPTY_TRASH_DAYS":    "7",
		"DISABLE_WP_CRON":     "true",
		"WP_CACHE":            "true",
		"DISALLOW_FILE_EDIT":  "true",
		"FORCE_SSL_ADMIN":     "true",
	}

	for key, value := range configs {
		wp.run("config", "set", key, value, "--raw")
	}

	// DragonflyDB/Redis settings
	redisConfigs := map[string]string{
		"WP_REDIS_HOST":     "'127.0.0.1'",
		"WP_REDIS_PORT":     "6379",
		"WP_REDIS_DATABASE": "0",
	}

	for key, value := range redisConfigs {
		wp.run("config", "set", key, value, "--raw")
	}

	return nil
}

// InstallPlugin installs and activates a plugin
func (wp *WordPress) InstallPlugin(slug string) error {
	return wp.run("plugin", "install", slug, "--activate")
}

// InstallTheme installs and activates a theme
func (wp *WordPress) InstallTheme(slug string) error {
	if err := wp.run("theme", "install", slug); err != nil {
		return err
	}
	return wp.run("theme", "activate", slug)
}

// SetupObjectCache installs Redis object cache
func (wp *WordPress) SetupObjectCache() error {
	// Install Redis Cache plugin
	if err := wp.InstallPlugin("redis-cache"); err != nil {
		return fmt.Errorf("failed to install redis-cache: %w", err)
	}

	// Enable Redis
	return wp.run("redis", "enable")
}

// UpdateCore updates WordPress core
func (wp *WordPress) UpdateCore() error {
	return wp.run("core", "update")
}

// UpdatePlugins updates all plugins
func (wp *WordPress) UpdatePlugins() error {
	return wp.run("plugin", "update", "--all")
}

// UpdateThemes updates all themes
func (wp *WordPress) UpdateThemes() error {
	return wp.run("theme", "update", "--all")
}

// UpdateAll updates core, plugins, and themes
func (wp *WordPress) UpdateAll() error {
	wp.UpdateCore()
	wp.UpdatePlugins()
	wp.UpdateThemes()
	return nil
}

// Harden applies security hardening
func (wp *WordPress) Harden() error {
	// Remove default themes
	wp.run("theme", "delete", "twentytwentytwo")
	wp.run("theme", "delete", "twentytwentythree")

	// Remove Hello Dolly and Akismet if not activated
	wp.run("plugin", "delete", "hello")
	wp.run("plugin", "delete", "akismet")

	// Disable XML-RPC
	wp.run("config", "set", "XMLRPC_REQUEST", "false", "--raw")

	// Set secure file permissions
	publicDir := filepath.Join(wp.Path, "public")
	exec.Command("find", publicDir, "-type", "d", "-exec", "chmod", "755", "{}", ";").Run()
	exec.Command("find", publicDir, "-type", "f", "-exec", "chmod", "644", "{}", ";").Run()
	exec.Command("chmod", "400", filepath.Join(publicDir, "wp-config.php")).Run()

	return nil
}

// EnableMultisite enables WordPress Multisite
func (wp *WordPress) EnableMultisite(subdomain bool) error {
	mode := "subdirectory"
	if subdomain {
		mode = "subdomain"
	}
	return wp.run("core", "multisite-convert", "--"+mode)
}

// ListPlugins returns installed plugins
func (wp *WordPress) ListPlugins() ([]Plugin, error) {
	out, err := wp.output("plugin", "list", "--format=csv", "--fields=name,status,version")
	if err != nil {
		return nil, err
	}
	return parsePluginList(out), nil
}

// ListThemes returns installed themes
func (wp *WordPress) ListThemes() ([]Theme, error) {
	out, err := wp.output("theme", "list", "--format=csv", "--fields=name,status,version")
	if err != nil {
		return nil, err
	}
	return parseThemeList(out), nil
}

// SearchReplace performs database search and replace
func (wp *WordPress) SearchReplace(from, to string) error {
	return wp.run("search-replace", from, to, "--all-tables")
}

// ExportDB exports the database
func (wp *WordPress) ExportDB(outputPath string) error {
	return wp.run("db", "export", outputPath)
}

// ImportDB imports a database
func (wp *WordPress) ImportDB(inputPath string) error {
	return wp.run("db", "import", inputPath)
}

// OptimizeDB optimizes database tables
func (wp *WordPress) OptimizeDB() error {
	return wp.run("db", "optimize")
}

// FlushCache flushes all caches
func (wp *WordPress) FlushCache() error {
	wp.run("cache", "flush")
	wp.run("redis", "flush")
	return nil
}

// run executes a WP-CLI command
func (wp *WordPress) run(args ...string) error {
	args = append(args, "--path="+filepath.Join(wp.Path, "public"))
	cmd := exec.Command("wp", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// output runs a command and returns output
func (wp *WordPress) output(args ...string) (string, error) {
	args = append(args, "--path="+filepath.Join(wp.Path, "public"))
	out, err := exec.Command("wp", args...).Output()
	return string(out), err
}

// Plugin represents a WordPress plugin
type Plugin struct {
	Name    string
	Status  string
	Version string
}

// Theme represents a WordPress theme
type Theme struct {
	Name    string
	Status  string
	Version string
}

func parsePluginList(csv string) []Plugin {
	var plugins []Plugin
	lines := strings.Split(csv, "\n")
	for i, line := range lines {
		if i == 0 || line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) >= 3 {
			plugins = append(plugins, Plugin{
				Name:    parts[0],
				Status:  parts[1],
				Version: parts[2],
			})
		}
	}
	return plugins
}

func parseThemeList(csv string) []Theme {
	var themes []Theme
	lines := strings.Split(csv, "\n")
	for i, line := range lines {
		if i == 0 || line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) >= 3 {
			themes = append(themes, Theme{
				Name:    parts[0],
				Status:  parts[1],
				Version: parts[2],
			})
		}
	}
	return themes
}
