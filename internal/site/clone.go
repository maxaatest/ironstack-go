package site

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Clone creates a copy of a site
func (m *Manager) Clone(sourceDomain, targetDomain string) error {
	sourcePath := filepath.Join(m.WebRoot, sourceDomain)
	targetPath := filepath.Join(m.WebRoot, targetDomain)

	// Copy files
	if err := exec.Command("cp", "-r", sourcePath, targetPath).Run(); err != nil {
		return fmt.Errorf("copy failed: %w", err)
	}

	// Create new database
	targetSite := &Site{
		Domain:     targetDomain,
		Path:       targetPath,
		UseVarnish: true,
	}
	
	dbPass, err := m.createDatabase(targetSite)
	if err != nil {
		return fmt.Errorf("database creation failed: %w", err)
	}

	// Export source database
	tempSQL := "/tmp/clone_" + sourceDomain + ".sql"
	exec.Command("wp", "db", "export", tempSQL, "--path="+sourcePath+"/public").Run()

	// Import to target database
	exec.Command("wp", "db", "import", tempSQL, "--path="+targetPath+"/public").Run()
	os.Remove(tempSQL)

	// Update wp-config with new database credentials
	wpConfigPath := filepath.Join(targetPath, "public", "wp-config.php")
	content, _ := os.ReadFile(wpConfigPath)
	
	newContent := string(content)
	newContent = replaceConfigValue(newContent, "DB_NAME", targetSite.DBName)
	newContent = replaceConfigValue(newContent, "DB_USER", targetSite.DBUser)
	newContent = replaceConfigValue(newContent, "DB_PASSWORD", dbPass)
	
	os.WriteFile(wpConfigPath, []byte(newContent), 0644)

	// Search-replace URLs in database
	exec.Command("wp", "search-replace", 
		"https://"+sourceDomain, 
		"https://"+targetDomain, 
		"--all-tables",
		"--path="+targetPath+"/public",
	).Run()

	exec.Command("wp", "search-replace", 
		"http://"+sourceDomain, 
		"https://"+targetDomain, 
		"--all-tables",
		"--path="+targetPath+"/public",
	).Run()

	// Generate Caddy config for new domain
	m.CaddyConf.AddSite(targetDomain, true)

	// Set permissions
	exec.Command("chown", "-R", "www-data:www-data", targetPath).Run()

	// Reload Caddy
	exec.Command("systemctl", "reload", "caddy").Run()

	return nil
}

// CreateStaging creates a staging environment
func (m *Manager) CreateStaging(domain string) error {
	stagingDomain := "staging." + domain
	return m.Clone(domain, stagingDomain)
}

// PushToProduction pushes staging to production
func (m *Manager) PushToProduction(domain string) error {
	stagingDomain := "staging." + domain
	stagingPath := filepath.Join(m.WebRoot, stagingDomain)
	prodPath := filepath.Join(m.WebRoot, domain)

	// Backup production first
	backupPath := fmt.Sprintf("/backups/%s/pre-push_%s.tar.gz", domain, time.Now().Format("2006-01-02_15-04-05"))
	os.MkdirAll(filepath.Dir(backupPath), 0755)
	exec.Command("tar", "-czf", backupPath, "-C", m.WebRoot, domain).Run()

	// Export staging database
	tempSQL := "/tmp/staging_" + domain + ".sql"
	exec.Command("wp", "db", "export", tempSQL, "--path="+stagingPath+"/public").Run()

	// Sync files (excluding wp-config.php)
	exec.Command("rsync", "-av", "--delete",
		"--exclude=wp-config.php",
		"--exclude=.htaccess",
		stagingPath+"/public/",
		prodPath+"/public/",
	).Run()

	// Import database to production
	exec.Command("wp", "db", "import", tempSQL, "--path="+prodPath+"/public").Run()
	os.Remove(tempSQL)

	// Search-replace URLs
	exec.Command("wp", "search-replace",
		"https://"+stagingDomain,
		"https://"+domain,
		"--all-tables",
		"--path="+prodPath+"/public",
	).Run()

	// Flush caches
	exec.Command("wp", "cache", "flush", "--path="+prodPath+"/public").Run()

	return nil
}

// ListDomains returns all domains with their status
func (m *Manager) ListDomains() ([]DomainInfo, error) {
	entries, err := os.ReadDir(m.WebRoot)
	if err != nil {
		return nil, err
	}

	var domains []DomainInfo
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		
		info := DomainInfo{
			Domain: e.Name(),
			Path:   filepath.Join(m.WebRoot, e.Name()),
		}
		
		// Check if WordPress is installed
		wpConfig := filepath.Join(info.Path, "public", "wp-config.php")
		if _, err := os.Stat(wpConfig); err == nil {
			info.HasWordPress = true
		}
		
		// Check SSL status
		info.HasSSL = m.checkSSL(e.Name())
		
		// Check if staging
		info.IsStaging = strings.HasPrefix(e.Name(), "staging.")
		
		domains = append(domains, info)
	}
	
	return domains, nil
}

// DomainInfo contains domain information
type DomainInfo struct {
	Domain       string
	Path         string
	HasWordPress bool
	HasSSL       bool
	IsStaging    bool
}

// checkSSL checks if domain has valid SSL
func (m *Manager) checkSSL(domain string) bool {
	err := exec.Command("curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", 
		"--connect-timeout", "5", "https://"+domain).Run()
	return err == nil
}

// AddDomain adds a new domain (alias) to existing site
func (m *Manager) AddDomain(siteDomain, newDomain string) error {
	sitePath := filepath.Join(m.WebRoot, siteDomain)
	
	// Create symlink
	linkPath := filepath.Join(m.WebRoot, newDomain)
	if err := os.Symlink(sitePath, linkPath); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}
	
	// Generate Caddy config
	m.CaddyConf.AddSite(newDomain, true)
	
	// Reload Caddy
	exec.Command("systemctl", "reload", "caddy").Run()
	
	return nil
}

// RemoveDomain removes a domain alias
func (m *Manager) RemoveDomain(domain string) error {
	linkPath := filepath.Join(m.WebRoot, domain)
	
	// Check if it's a symlink (alias) or real site
	fi, err := os.Lstat(linkPath)
	if err != nil {
		return err
	}
	
	if fi.Mode()&os.ModeSymlink != 0 {
		// It's a symlink, safe to remove
		os.Remove(linkPath)
	} else {
		// It's a real directory, delete completely
		os.RemoveAll(linkPath)
	}
	
	// Remove Caddy config
	os.Remove(filepath.Join("/etc/caddy/sites", domain+".conf"))
	
	// Reload Caddy
	exec.Command("systemctl", "reload", "caddy").Run()
	
	return nil
}

// SetMaintenanceMode enables/disables maintenance mode
func (m *Manager) SetMaintenanceMode(domain string, enabled bool) error {
	maintenanceFile := filepath.Join(m.WebRoot, domain, "public", ".maintenance")
	
	if enabled {
		content := `<?php $upgrading = time(); ?>`
		return os.WriteFile(maintenanceFile, []byte(content), 0644)
	}
	
	return os.Remove(maintenanceFile)
}

func replaceConfigValue(content, key, value string) string {
	// Replace define('KEY', 'old_value') with define('KEY', 'new_value')
	pattern := fmt.Sprintf(`define\s*\(\s*'%s'\s*,\s*'[^']*'\s*\)`, key)
	replacement := fmt.Sprintf("define('%s', '%s')", key, value)
	return strings.ReplaceAll(content, pattern, replacement)
}
