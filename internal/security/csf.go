package security

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// CSF manages ConfigServer Firewall
type CSF struct{}

// NewCSF creates a CSF manager
func NewCSF() *CSF {
	return &CSF{}
}

// Install installs CSF firewall
func (c *CSF) Install() error {
	commands := []string{
		"cd /usr/src && curl -O https://download.configserver.com/csf.tgz",
		"cd /usr/src && tar -xzf csf.tgz",
		"cd /usr/src/csf && sh install.sh",
	}
	for _, cmd := range commands {
		if err := exec.Command("sh", "-c", cmd).Run(); err != nil {
			return fmt.Errorf("install failed: %w", err)
		}
	}
	return nil
}

// Enable enables and starts CSF
func (c *CSF) Enable() error {
	// Disable testing mode
	c.SetConfig("TESTING", "0")
	return exec.Command("csf", "-e").Run()
}

// Disable disables CSF
func (c *CSF) Disable() error {
	return exec.Command("csf", "-x").Run()
}

// Restart restarts CSF
func (c *CSF) Restart() error {
	return exec.Command("csf", "-r").Run()
}

// Status returns CSF status
func (c *CSF) Status() (bool, error) {
	err := exec.Command("csf", "-l").Run()
	return err == nil, nil
}

// AllowIP adds an IP to whitelist
func (c *CSF) AllowIP(ip, comment string) error {
	return exec.Command("csf", "-a", ip, comment).Run()
}

// DenyIP adds an IP to blacklist
func (c *CSF) DenyIP(ip, comment string) error {
	return exec.Command("csf", "-d", ip, comment).Run()
}

// RemoveIP removes an IP from both lists
func (c *CSF) RemoveIP(ip string) error {
	exec.Command("csf", "-ar", ip).Run()
	exec.Command("csf", "-dr", ip).Run()
	return nil
}

// TempBlockIP temporarily blocks an IP
func (c *CSF) TempBlockIP(ip string, seconds int, comment string) error {
	return exec.Command("csf", "-td", ip, fmt.Sprintf("%d", seconds), comment).Run()
}

// OpenPort opens a TCP port
func (c *CSF) OpenPort(port int) error {
	// Add to TCP_IN and TCP_OUT
	c.AddToConfig("TCP_IN", fmt.Sprintf("%d", port))
	c.AddToConfig("TCP_OUT", fmt.Sprintf("%d", port))
	return c.Restart()
}

// ClosePort closes a TCP port
func (c *CSF) ClosePort(port int) error {
	c.RemoveFromConfig("TCP_IN", fmt.Sprintf("%d", port))
	c.RemoveFromConfig("TCP_OUT", fmt.Sprintf("%d", port))
	return c.Restart()
}

// OpenUDPPort opens a UDP port
func (c *CSF) OpenUDPPort(port int) error {
	c.AddToConfig("UDP_IN", fmt.Sprintf("%d", port))
	c.AddToConfig("UDP_OUT", fmt.Sprintf("%d", port))
	return c.Restart()
}

// GetBlockedIPs returns list of blocked IPs
func (c *CSF) GetBlockedIPs() ([]string, error) {
	out, err := os.ReadFile("/etc/csf/csf.deny")
	if err != nil {
		return nil, err
	}
	
	var ips []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			parts := strings.Split(line, " ")
			if len(parts) > 0 {
				ips = append(ips, parts[0])
			}
		}
	}
	return ips, nil
}

// GetAllowedIPs returns list of allowed IPs
func (c *CSF) GetAllowedIPs() ([]string, error) {
	out, err := os.ReadFile("/etc/csf/csf.allow")
	if err != nil {
		return nil, err
	}
	
	var ips []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			parts := strings.Split(line, " ")
			if len(parts) > 0 {
				ips = append(ips, parts[0])
			}
		}
	}
	return ips, nil
}

// SetConfig sets a configuration value
func (c *CSF) SetConfig(key, value string) error {
	return exec.Command("sed", "-i", fmt.Sprintf(`s/^%s = .*/%s = "%s"/`, key, key, value), "/etc/csf/csf.conf").Run()
}

// AddToConfig adds a value to a comma-separated config
func (c *CSF) AddToConfig(key, value string) error {
	// Read current config
	out, err := exec.Command("grep", fmt.Sprintf("^%s =", key), "/etc/csf/csf.conf").Output()
	if err != nil {
		return err
	}
	
	current := strings.TrimSpace(string(out))
	current = strings.Split(current, "=")[1]
	current = strings.Trim(current, `" `)
	
	if !strings.Contains(current, value) {
		if current != "" {
			current += ","
		}
		current += value
		return c.SetConfig(key, current)
	}
	return nil
}

// RemoveFromConfig removes a value from comma-separated config
func (c *CSF) RemoveFromConfig(key, value string) error {
	out, err := exec.Command("grep", fmt.Sprintf("^%s =", key), "/etc/csf/csf.conf").Output()
	if err != nil {
		return err
	}
	
	current := strings.TrimSpace(string(out))
	current = strings.Split(current, "=")[1]
	current = strings.Trim(current, `" `)
	
	parts := strings.Split(current, ",")
	var newParts []string
	for _, p := range parts {
		if strings.TrimSpace(p) != value {
			newParts = append(newParts, p)
		}
	}
	
	return c.SetConfig(key, strings.Join(newParts, ","))
}

// ConfigureForWordPress applies WordPress-optimized settings
func (c *CSF) ConfigureForWordPress() error {
	// Essential ports for WordPress
	essentialPorts := "80,443,22,25,53,587,993,995"
	
	c.SetConfig("TCP_IN", essentialPorts)
	c.SetConfig("TCP_OUT", "1:65535")
	c.SetConfig("UDP_IN", "53")
	c.SetConfig("UDP_OUT", "53,123,6277,6672")
	
	// Security settings
	c.SetConfig("SYNFLOOD", "1")
	c.SetConfig("SYNFLOOD_RATE", "75/s")
	c.SetConfig("SYNFLOOD_BURST", "25")
	c.SetConfig("PORTFLOOD", "22;tcp;5;300,80;tcp;20;5,443;tcp;20;5")
	c.SetConfig("CONNLIMIT", "22;5,80;50,443;50")
	c.SetConfig("CT_LIMIT", "200")
	c.SetConfig("LF_TRIGGER", "0")
	
	return c.Restart()
}
