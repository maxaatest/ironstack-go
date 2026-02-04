package security

import (
	"os"
	"os/exec"
)

// Security provides unified security management
type Security struct {
	CSF      *CSF
	Fail2ban *Fail2ban
}

// New creates a unified security manager
func New() *Security {
	return &Security{
		CSF:      NewCSF(),
		Fail2ban: NewFail2ban(),
	}
}

// InstallAll installs all security components
func (s *Security) InstallAll() error {
	if err := s.CSF.Install(); err != nil {
		return err
	}
	if err := s.Fail2ban.Install(); err != nil {
		return err
	}
	return nil
}

// ConfigureForWordPress applies WordPress security settings
func (s *Security) ConfigureForWordPress() error {
	// Configure CSF
	s.CSF.ConfigureForWordPress()
	s.CSF.Enable()
	
	// Configure Fail2ban
	s.Fail2ban.CreateWordPressJails()
	s.Fail2ban.ConfigureSSH()
	s.Fail2ban.Start()
	
	return nil
}

// HardenServer applies security hardening
func (s *Security) HardenServer() error {
	// Disable root SSH login
	exec.Command("sed", "-i", "s/PermitRootLogin yes/PermitRootLogin no/", "/etc/ssh/sshd_config").Run()
	
	// Disable password authentication (use keys only)
	// exec.Command("sed", "-i", "s/PasswordAuthentication yes/PasswordAuthentication no/", "/etc/ssh/sshd_config").Run()
	
	// Enable automatic security updates
	exec.Command("apt-get", "install", "-y", "unattended-upgrades").Run()
	exec.Command("dpkg-reconfigure", "-plow", "unattended-upgrades").Run()
	
	// Disable unused services
	services := []string{"cups", "avahi-daemon", "bluetooth"}
	for _, svc := range services {
		exec.Command("systemctl", "disable", svc).Run()
		exec.Command("systemctl", "stop", svc).Run()
	}
	
	// Set secure permissions on sensitive files
	exec.Command("chmod", "600", "/etc/shadow").Run()
	exec.Command("chmod", "600", "/etc/gshadow").Run()
	
	return nil
}

// BlockCountries blocks traffic from specific countries
func (s *Security) BlockCountries(countryCodes []string) error {
	for _, code := range countryCodes {
		s.CSF.AddToConfig("CC_DENY", code)
	}
	return s.CSF.Restart()
}

// AllowCountries allows traffic only from specific countries
func (s *Security) AllowCountries(countryCodes []string) error {
	s.CSF.SetConfig("CC_DENY", "")
	for _, code := range countryCodes {
		s.CSF.AddToConfig("CC_ALLOW", code)
	}
	s.CSF.SetConfig("CC_ALLOW_FILTER", "1")
	return s.CSF.Restart()
}

// BlockIP blocks an IP in both CSF and Fail2ban
func (s *Security) BlockIP(ip, reason string) error {
	s.CSF.DenyIP(ip, reason)
	s.Fail2ban.BanIP("sshd", ip)
	return nil
}

// UnblockIP unblocks an IP from both CSF and Fail2ban
func (s *Security) UnblockIP(ip string) error {
	s.CSF.RemoveIP(ip)
	s.Fail2ban.UnbanIP("sshd", ip)
	return nil
}

// Status returns security service status
func (s *Security) Status() map[string]bool {
	status := make(map[string]bool)
	status["csf"], _ = s.CSF.Status()
	status["fail2ban"], _ = s.Fail2ban.Status()
	return status
}

// GenerateSecurityReport generates a security report
func (s *Security) GenerateSecurityReport() string {
	report := "=== Security Report ===\n\n"
	
	// CSF Status
	csfActive, _ := s.CSF.Status()
	report += "CSF Firewall: "
	if csfActive {
		report += "ACTIVE ✓\n"
	} else {
		report += "INACTIVE ✗\n"
	}
	
	// Fail2ban Status
	f2bActive, _ := s.Fail2ban.Status()
	report += "Fail2ban: "
	if f2bActive {
		report += "ACTIVE ✓\n"
	} else {
		report += "INACTIVE ✗\n"
	}
	
	// Blocked IPs
	blockedIPs, _ := s.CSF.GetBlockedIPs()
	report += fmt.Sprintf("Blocked IPs: %d\n", len(blockedIPs))
	
	// Active jails
	jails, _ := s.Fail2ban.GetJails()
	report += fmt.Sprintf("Active Jails: %d\n", len(jails))
	
	return report
}

// WordPressSecurityRules creates WordPress-specific htaccess rules
func (s *Security) WordPressSecurityRules(sitePath string) error {
	htaccess := `# IronStack Security Rules

# Disable directory browsing
Options -Indexes

# Block access to sensitive files
<FilesMatch "^(wp-config\.php|\.htaccess|readme\.html|license\.txt)$">
    Order Allow,Deny
    Deny from all
</FilesMatch>

# Block access to includes folder
<IfModule mod_rewrite.c>
    RewriteEngine On
    RewriteBase /
    RewriteRule ^wp-admin/includes/ - [F,L]
    RewriteRule !^wp-includes/ - [S=3]
    RewriteRule ^wp-includes/[^/]+\.php$ - [F,L]
    RewriteRule ^wp-includes/js/tinymce/langs/.+\.php - [F,L]
    RewriteRule ^wp-includes/theme-compat/ - [F,L]
</IfModule>

# Block author scans
<IfModule mod_rewrite.c>
    RewriteEngine On
    RewriteBase /
    RewriteCond %{QUERY_STRING} (author=\d+) [NC]
    RewriteRule .* - [F]
</IfModule>

# Limit file uploads
<IfModule mod_php.c>
    php_value upload_max_filesize 64M
    php_value post_max_size 64M
    php_value max_execution_time 300
    php_value max_input_time 300
</IfModule>
`
	return os.WriteFile(sitePath+"/public/.htaccess-security", []byte(htaccess), 0644)
}
