package modules

import (
	"os"
	"os/exec"
)

// Security manages CSF and Fail2ban
type Security struct{}

func NewSecurity() *Security {
	return &Security{}
}

func (s *Security) InstallCSF() error {
	cmd := `cd /usr/src &&
		wget https://download.configserver.com/csf.tgz &&
		tar -xzf csf.tgz &&
		cd csf &&
		sh install.sh`
	return exec.Command("sh", "-c", cmd).Run()
}

func (s *Security) InstallFail2ban() error {
	if err := exec.Command("apt-get", "install", "-y", "fail2ban").Run(); err != nil {
		return err
	}

	// WordPress jail config
	jailConfig := `
[wordpress]
enabled = true
filter = wordpress
logpath = /var/log/caddy/access.log
maxretry = 5
bantime = 3600
findtime = 600

[sshd]
enabled = true
port = ssh
filter = sshd
logpath = /var/log/auth.log
maxretry = 3
bantime = 3600
`
	if err := os.WriteFile("/etc/fail2ban/jail.local", []byte(jailConfig), 0644); err != nil {
		return err
	}

	// WordPress filter
	wpFilter := `
[Definition]
failregex = ^<HOST> .* "POST /wp-login.php
            ^<HOST> .* "POST /xmlrpc.php
ignoreregex =
`
	if err := os.WriteFile("/etc/fail2ban/filter.d/wordpress.conf", []byte(wpFilter), 0644); err != nil {
		return err
	}

	return exec.Command("systemctl", "restart", "fail2ban").Run()
}

func (s *Security) BlockIP(ip string) error {
	return exec.Command("csf", "-d", ip).Run()
}

func (s *Security) UnblockIP(ip string) error {
	return exec.Command("csf", "-dr", ip).Run()
}

func (s *Security) AllowPort(port string) error {
	if err := exec.Command("csf", "-a", port).Run(); err != nil {
		return err
	}
	return exec.Command("csf", "-r").Run()
}
