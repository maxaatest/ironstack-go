package site

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// SSL manages SSL certificates via Caddy
type SSL struct{}

// NewSSL creates an SSL manager
func NewSSL() *SSL {
	return &SSL{}
}

// CertInfo contains SSL certificate information
type CertInfo struct {
	Domain     string
	Issuer     string
	ValidFrom  time.Time
	ValidUntil time.Time
	DaysLeft   int
	AutoRenew  bool
}

// GetCertInfo returns SSL certificate information for a domain
func (s *SSL) GetCertInfo(domain string) (*CertInfo, error) {
	// Use openssl to check certificate
	out, err := exec.Command("sh", "-c", 
		fmt.Sprintf("echo | openssl s_client -servername %s -connect %s:443 2>/dev/null | openssl x509 -noout -dates -issuer", domain, domain),
	).Output()
	
	if err != nil {
		return nil, fmt.Errorf("failed to get cert info: %w", err)
	}
	
	info := &CertInfo{Domain: domain, AutoRenew: true}
	
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "notBefore=") {
			dateStr := strings.TrimPrefix(line, "notBefore=")
			if t, err := time.Parse("Jan  2 15:04:05 2006 GMT", dateStr); err == nil {
				info.ValidFrom = t
			}
		} else if strings.HasPrefix(line, "notAfter=") {
			dateStr := strings.TrimPrefix(line, "notAfter=")
			if t, err := time.Parse("Jan  2 15:04:05 2006 GMT", dateStr); err == nil {
				info.ValidUntil = t
				info.DaysLeft = int(time.Until(t).Hours() / 24)
			}
		} else if strings.HasPrefix(line, "issuer=") {
			info.Issuer = strings.TrimPrefix(line, "issuer=")
		}
	}
	
	return info, nil
}

// ForceCertRenewal forces certificate renewal
func (s *SSL) ForceCertRenewal(domain string) error {
	// Caddy handles auto-renewal, but we can force it
	return exec.Command("caddy", "reload", "--config", "/etc/caddy/Caddyfile").Run()
}

// ListCertificates lists all SSL certificates
func (s *SSL) ListCertificates() ([]CertInfo, error) {
	// Get certificates from Caddy's data directory
	out, err := exec.Command("find", "/var/lib/caddy/.local/share/caddy/certificates", 
		"-name", "*.crt", "-type", "f").Output()
	
	if err != nil {
		return nil, err
	}
	
	var certs []CertInfo
	files := strings.Split(strings.TrimSpace(string(out)), "\n")
	
	for _, file := range files {
		if file == "" {
			continue
		}
		
		// Extract domain from path
		parts := strings.Split(file, "/")
		if len(parts) > 0 {
			domain := strings.TrimSuffix(parts[len(parts)-1], ".crt")
			if info, err := s.GetCertInfo(domain); err == nil {
				certs = append(certs, *info)
			}
		}
	}
	
	return certs, nil
}

// CheckExpiring returns certificates expiring within days
func (s *SSL) CheckExpiring(days int) ([]CertInfo, error) {
	certs, err := s.ListCertificates()
	if err != nil {
		return nil, err
	}
	
	var expiring []CertInfo
	for _, cert := range certs {
		if cert.DaysLeft <= days {
			expiring = append(expiring, cert)
		}
	}
	
	return expiring, nil
}

// TestSSL tests SSL configuration for a domain
func (s *SSL) TestSSL(domain string) (*SSLTestResult, error) {
	result := &SSLTestResult{Domain: domain}
	
	// Check HTTPS accessibility
	out, _ := exec.Command("curl", "-s", "-o", "/dev/null", "-w", "%{http_code}",
		"--connect-timeout", "10", "https://"+domain).Output()
	result.HTTPSAccessible = strings.TrimSpace(string(out)) == "200"
	
	// Check HTTP to HTTPS redirect
	out, _ = exec.Command("curl", "-s", "-o", "/dev/null", "-w", "%{redirect_url}",
		"--connect-timeout", "10", "-L", "http://"+domain).Output()
	result.HTTPRedirect = strings.HasPrefix(strings.TrimSpace(string(out)), "https://")
	
	// Check certificate validity
	err := exec.Command("sh", "-c",
		fmt.Sprintf("echo | openssl s_client -servername %s -connect %s:443 2>/dev/null | openssl x509 -noout -checkend 0", domain, domain),
	).Run()
	result.ValidCert = err == nil
	
	// Check HSTS header
	out, _ = exec.Command("curl", "-s", "-I", "https://"+domain).Output()
	result.HSTS = strings.Contains(string(out), "strict-transport-security")
	
	return result, nil
}

// SSLTestResult contains SSL test results
type SSLTestResult struct {
	Domain          string
	HTTPSAccessible bool
	HTTPRedirect    bool
	ValidCert       bool
	HSTS            bool
}

// Score returns SSL quality score
func (r *SSLTestResult) Score() int {
	score := 0
	if r.HTTPSAccessible {
		score += 25
	}
	if r.HTTPRedirect {
		score += 25
	}
	if r.ValidCert {
		score += 25
	}
	if r.HSTS {
		score += 25
	}
	return score
}

// ToJSON returns JSON representation
func (r *SSLTestResult) ToJSON() string {
	data, _ := json.MarshalIndent(r, "", "  ")
	return string(data)
}
