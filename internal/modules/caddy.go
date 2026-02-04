package modules

import (
	"fmt"
	"os/exec"
)

// Caddy manages Caddy web server
type Caddy struct{}

func NewCaddy() *Caddy {
	return &Caddy{}
}

func (c *Caddy) Install() error {
	cmd := exec.Command("sh", "-c", `
		apt-get update && 
		apt-get install -y debian-keyring debian-archive-keyring apt-transport-https &&
		curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg &&
		curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | tee /etc/apt/sources.list.d/caddy-stable.list &&
		apt-get update &&
		apt-get install -y caddy
	`)
	return cmd.Run()
}

func (c *Caddy) AddSite(domain string) error {
	config := fmt.Sprintf(`
%s {
    root * /var/www/%s
    php_fastcgi 127.0.0.1:9000
    file_server
    encode gzip
    
    @static {
        path *.css *.js *.ico *.gif *.jpg *.jpeg *.png *.svg *.woff2
    }
    header @static Cache-Control "public, max-age=31536000"
}
`, domain, domain)

	return exec.Command("sh", "-c", fmt.Sprintf("echo '%s' >> /etc/caddy/Caddyfile", config)).Run()
}

func (c *Caddy) Reload() error {
	return exec.Command("systemctl", "reload", "caddy").Run()
}

func (c *Caddy) Status() (string, error) {
	out, err := exec.Command("systemctl", "is-active", "caddy").Output()
	return string(out), err
}
