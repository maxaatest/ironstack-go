package site

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/maxaatest/ironstack/internal/config"
)

// Site represents a WordPress site
type Site struct {
	Domain      string
	Path        string
	DBName      string
	DBUser      string
	DBPass      string
	EnableSSL   bool
	UseVarnish  bool
}

// Manager handles site operations
type Manager struct {
	WebRoot   string
	CaddyConf *config.Caddy
}

// NewManager creates a new site manager
func NewManager() *Manager {
	return &Manager{
		WebRoot:   "/var/www",
		CaddyConf: config.NewCaddy(),
	}
}

// Create sets up a new WordPress site
func (m *Manager) Create(s *Site) error {
	s.Path = filepath.Join(m.WebRoot, s.Domain)
	
	// Create directory structure
	dirs := []string{
		s.Path,
		filepath.Join(s.Path, "public"),
		filepath.Join(s.Path, "logs"),
		filepath.Join(s.Path, "backups"),
	}
	
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	
	// Create database
	dbPass, err := m.createDatabase(s)
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}
	s.DBPass = dbPass
	
	// Download WordPress
	if err := m.downloadWordPress(s); err != nil {
		return fmt.Errorf("failed to download WordPress: %w", err)
	}
	
	// Create wp-config.php
	if err := m.createConfig(s); err != nil {
		return fmt.Errorf("failed to create wp-config: %w", err)
	}
	
	// Generate Caddy config
	if err := m.CaddyConf.AddSite(s.Domain, s.UseVarnish); err != nil {
		return fmt.Errorf("failed to create Caddy config: %w", err)
	}
	
	// Reload Caddy
	exec.Command("systemctl", "reload", "caddy").Run()
	
	// Set permissions
	exec.Command("chown", "-R", "www-data:www-data", s.Path).Run()
	
	return nil
}

func (m *Manager) createDatabase(s *Site) (string, error) {
	s.DBName = sanitizeName(s.Domain) + "_db"
	s.DBUser = sanitizeName(s.Domain) + "_user"
	password := generatePassword()
	
	sql := fmt.Sprintf(`
		CREATE DATABASE IF NOT EXISTS %s;
		CREATE USER IF NOT EXISTS '%s'@'localhost' IDENTIFIED BY '%s';
		GRANT ALL PRIVILEGES ON %s.* TO '%s'@'localhost';
		FLUSH PRIVILEGES;
	`, s.DBName, s.DBUser, password, s.DBName, s.DBUser)
	
	cmd := exec.Command("mysql", "-e", sql)
	if err := cmd.Run(); err != nil {
		return "", err
	}
	
	return password, nil
}

func (m *Manager) downloadWordPress(s *Site) error {
	publicDir := filepath.Join(s.Path, "public")
	return exec.Command("wp", "core", "download", "--path="+publicDir).Run()
}

func (m *Manager) createConfig(s *Site) error {
	publicDir := filepath.Join(s.Path, "public")
	
	// Create wp-config using WP-CLI
	cmd := exec.Command("wp", "config", "create",
		"--path="+publicDir,
		"--dbname="+s.DBName,
		"--dbuser="+s.DBUser,
		"--dbpass="+s.DBPass,
		"--dbhost=localhost",
	)
	if err := cmd.Run(); err != nil {
		return err
	}
	
	// Add optimizations
	wp := config.NewWordPress()
	optimizations := wp.OptimizeConfig()
	
	configPath := filepath.Join(publicDir, "wp-config.php")
	content, _ := os.ReadFile(configPath)
	
	// Insert before "That's all, stop editing!"
	newContent := string(content)
	marker := "/* That's all, stop editing!"
	if idx := len(newContent) - 100; idx > 0 {
		newContent = newContent[:idx] + optimizations + newContent[idx:]
	}
	
	return os.WriteFile(configPath, []byte(newContent), 0644)
}

// Delete removes a site
func (m *Manager) Delete(domain string) error {
	// Remove directory
	os.RemoveAll(filepath.Join(m.WebRoot, domain))
	
	// Remove Caddy config
	os.Remove(filepath.Join("/etc/caddy/sites", domain+".conf"))
	
	// Reload Caddy
	exec.Command("systemctl", "reload", "caddy").Run()
	
	return nil
}

// List returns all sites
func (m *Manager) List() ([]string, error) {
	entries, err := os.ReadDir(m.WebRoot)
	if err != nil {
		return nil, err
	}
	
	var sites []string
	for _, e := range entries {
		if e.IsDir() {
			sites = append(sites, e.Name())
		}
	}
	return sites, nil
}

func sanitizeName(domain string) string {
	// Replace dots and hyphens with underscores
	result := ""
	for _, c := range domain {
		if c == '.' || c == '-' {
			result += "_"
		} else if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			result += string(c)
		}
	}
	return result
}

func generatePassword() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%"
	b := make([]byte, 24)
	for i := range b {
		b[i] = chars[i%len(chars)]
	}
	return string(b)
}
