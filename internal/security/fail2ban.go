package security

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Fail2ban manages Fail2ban service
type Fail2ban struct {
	JailDir string
}

// NewFail2ban creates a Fail2ban manager
func NewFail2ban() *Fail2ban {
	return &Fail2ban{JailDir: "/etc/fail2ban/jail.d"}
}

// Install installs Fail2ban
func (f *Fail2ban) Install() error {
	if err := exec.Command("apt-get", "install", "-y", "fail2ban").Run(); err != nil {
		return err
	}
	return exec.Command("systemctl", "enable", "fail2ban").Run()
}

// Start starts Fail2ban
func (f *Fail2ban) Start() error {
	return exec.Command("systemctl", "start", "fail2ban").Run()
}

// Stop stops Fail2ban
func (f *Fail2ban) Stop() error {
	return exec.Command("systemctl", "stop", "fail2ban").Run()
}

// Restart restarts Fail2ban
func (f *Fail2ban) Restart() error {
	return exec.Command("systemctl", "restart", "fail2ban").Run()
}

// Status returns Fail2ban status
func (f *Fail2ban) Status() (bool, error) {
	err := exec.Command("systemctl", "is-active", "--quiet", "fail2ban").Run()
	return err == nil, nil
}

// BanIP manually bans an IP in a jail
func (f *Fail2ban) BanIP(jail, ip string) error {
	return exec.Command("fail2ban-client", "set", jail, "banip", ip).Run()
}

// UnbanIP unbans an IP from a jail
func (f *Fail2ban) UnbanIP(jail, ip string) error {
	return exec.Command("fail2ban-client", "set", jail, "unbanip", ip).Run()
}

// GetBannedIPs returns banned IPs for a jail
func (f *Fail2ban) GetBannedIPs(jail string) ([]string, error) {
	out, err := exec.Command("fail2ban-client", "status", jail).Output()
	if err != nil {
		return nil, err
	}
	
	// Parse output to find banned IPs
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "Banned IP list:") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				ips := strings.Fields(parts[1])
				return ips, nil
			}
		}
	}
	return nil, nil
}

// GetJails returns list of active jails
func (f *Fail2ban) GetJails() ([]string, error) {
	out, err := exec.Command("fail2ban-client", "status").Output()
	if err != nil {
		return nil, err
	}
	
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "Jail list:") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				jails := strings.Split(strings.TrimSpace(parts[1]), ",")
				for i, j := range jails {
					jails[i] = strings.TrimSpace(j)
				}
				return jails, nil
			}
		}
	}
	return nil, nil
}

// CreateWordPressJails creates WordPress-specific jails
func (f *Fail2ban) CreateWordPressJails() error {
	os.MkdirAll(f.JailDir, 0755)
	
	// WordPress auth jail
	wpAuth := `[wordpress-auth]
enabled = true
port = http,https
filter = wordpress-auth
logpath = /var/log/caddy/*-access.log
maxretry = 5
bantime = 3600
findtime = 600
`
	if err := os.WriteFile(filepath.Join(f.JailDir, "wordpress.conf"), []byte(wpAuth), 0644); err != nil {
		return err
	}
	
	// Create filter
	filterDir := "/etc/fail2ban/filter.d"
	wpFilter := `[Definition]
failregex = ^<HOST> .* "POST /wp-login.php
            ^<HOST> .* "POST /xmlrpc.php
ignoreregex =
`
	if err := os.WriteFile(filepath.Join(filterDir, "wordpress-auth.conf"), []byte(wpFilter), 0644); err != nil {
		return err
	}
	
	return f.Restart()
}

// CreateWooCommerceJail creates WooCommerce-specific jail
func (f *Fail2ban) CreateWooCommerceJail() error {
	wooJail := `[woocommerce]
enabled = true
port = http,https
filter = woocommerce
logpath = /var/log/caddy/*-access.log
maxretry = 10
bantime = 1800
findtime = 300
`
	if err := os.WriteFile(filepath.Join(f.JailDir, "woocommerce.conf"), []byte(wooJail), 0644); err != nil {
		return err
	}
	
	filterDir := "/etc/fail2ban/filter.d"
	wooFilter := `[Definition]
failregex = ^<HOST> .* "POST /wp-admin/admin-ajax.php.*wc-ajax
            ^<HOST> .* "POST /.*add-to-cart
ignoreregex =
`
	if err := os.WriteFile(filepath.Join(filterDir, "woocommerce.conf"), []byte(wooFilter), 0644); err != nil {
		return err
	}
	
	return f.Restart()
}

// CreateBruteForceJail creates a general brute-force protection jail
func (f *Fail2ban) CreateBruteForceJail() error {
	bruteJail := `[http-brute]
enabled = true
port = http,https
filter = http-brute
logpath = /var/log/caddy/*-access.log
maxretry = 30
bantime = 600
findtime = 60
`
	if err := os.WriteFile(filepath.Join(f.JailDir, "http-brute.conf"), []byte(bruteJail), 0644); err != nil {
		return err
	}
	
	filterDir := "/etc/fail2ban/filter.d"
	bruteFilter := `[Definition]
failregex = ^<HOST> .* "(GET|POST|HEAD)
ignoreregex = \.(css|js|jpg|jpeg|png|gif|ico|svg|woff|woff2)
`
	if err := os.WriteFile(filepath.Join(filterDir, "http-brute.conf"), []byte(bruteFilter), 0644); err != nil {
		return err
	}
	
	return f.Restart()
}

// ConfigureSSH hardens SSH jail
func (f *Fail2ban) ConfigureSSH() error {
	sshJail := `[sshd]
enabled = true
port = ssh
filter = sshd
logpath = /var/log/auth.log
maxretry = 3
bantime = 86400
findtime = 600
`
	return os.WriteFile(filepath.Join(f.JailDir, "sshd.conf"), []byte(sshJail), 0644)
}

// GetJailStatus returns status of a specific jail
func (f *Fail2ban) GetJailStatus(jail string) (map[string]string, error) {
	out, err := exec.Command("fail2ban-client", "status", jail).Output()
	if err != nil {
		return nil, err
	}
	
	status := make(map[string]string)
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(strings.TrimPrefix(parts[0], "`-"))
				key = strings.TrimPrefix(key, "|- ")
				status[key] = strings.TrimSpace(parts[1])
			}
		}
	}
	return status, nil
}

// SetBanTime sets ban time for a jail
func (f *Fail2ban) SetBanTime(jail string, seconds int) error {
	return exec.Command("fail2ban-client", "set", jail, "bantime", fmt.Sprintf("%d", seconds)).Run()
}

// SetMaxRetry sets max retries for a jail
func (f *Fail2ban) SetMaxRetry(jail string, count int) error {
	return exec.Command("fail2ban-client", "set", jail, "maxretry", fmt.Sprintf("%d", count)).Run()
}
