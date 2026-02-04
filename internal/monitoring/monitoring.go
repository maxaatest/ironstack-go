package monitoring

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// GoAccess manages GoAccess analytics
type GoAccess struct {
	ReportDir string
}

// NewGoAccess creates a GoAccess manager
func NewGoAccess() *GoAccess {
	return &GoAccess{ReportDir: "/var/www/analytics"}
}

// Install installs GoAccess
func (g *GoAccess) Install() error {
	return exec.Command("apt-get", "install", "-y", "goaccess").Run()
}

// GenerateReport generates HTML report for a domain
func (g *GoAccess) GenerateReport(domain, logPath string) error {
	os.MkdirAll(g.ReportDir, 0755)
	
	outputPath := fmt.Sprintf("%s/%s.html", g.ReportDir, domain)
	
	return exec.Command("goaccess", logPath,
		"-o", outputPath,
		"--log-format=COMBINED",
		"--real-time-html",
		"--ws-url=wss://"+domain+":7890",
	).Run()
}

// StartRealtime starts real-time WebSocket server
func (g *GoAccess) StartRealtime(domain, logPath string) error {
	outputPath := fmt.Sprintf("%s/%s.html", g.ReportDir, domain)
	
	cmd := exec.Command("goaccess", logPath,
		"-o", outputPath,
		"--log-format=COMBINED",
		"--real-time-html",
		"--port=7890",
		"--daemonize",
	)
	return cmd.Start()
}

// GetStats returns parsed access log statistics
func (g *GoAccess) GetStats(logPath string) (*AccessStats, error) {
	out, err := exec.Command("goaccess", logPath,
		"--log-format=COMBINED",
		"-o", "json:-",
	).Output()
	
	if err != nil {
		return nil, err
	}
	
	var stats AccessStats
	if err := json.Unmarshal(out, &stats); err != nil {
		return nil, err
	}
	
	return &stats, nil
}

// AccessStats contains parsed log statistics
type AccessStats struct {
	TotalRequests int64   `json:"total_requests"`
	ValidRequests int64   `json:"valid_requests"`
	FailedRequests int64  `json:"failed_requests"`
	UniqueVisitors int64  `json:"unique_visitors"`
	Bandwidth      int64   `json:"bandwidth"`
	AvgTime        float64 `json:"avg_time_served"`
}

// Server monitors server resources
type Server struct{}

// NewServer creates a server monitor
func NewServer() *Server {
	return &Server{}
}

// Stats contains server statistics
type Stats struct {
	CPU        CPUStats
	Memory     MemoryStats
	Disk       DiskStats
	Load       LoadStats
	Uptime     string
	Hostname   string
	Processes  int
}

type CPUStats struct {
	Usage   float64
	Cores   int
	Model   string
}

type MemoryStats struct {
	Total     int64
	Used      int64
	Free      int64
	UsagePercent float64
}

type DiskStats struct {
	Total     int64
	Used      int64
	Free      int64
	UsagePercent float64
}

type LoadStats struct {
	Load1  float64
	Load5  float64
	Load15 float64
}

// GetStats retrieves current server statistics
func (s *Server) GetStats() (*Stats, error) {
	stats := &Stats{}
	
	// Hostname
	out, _ := exec.Command("hostname").Output()
	stats.Hostname = strings.TrimSpace(string(out))
	
	// Uptime
	out, _ = exec.Command("uptime", "-p").Output()
	stats.Uptime = strings.TrimPrefix(strings.TrimSpace(string(out)), "up ")
	
	// CPU
	out, _ = exec.Command("nproc").Output()
	stats.CPU.Cores, _ = strconv.Atoi(strings.TrimSpace(string(out)))
	
	out, _ = exec.Command("sh", "-c", "top -bn1 | grep 'Cpu(s)' | awk '{print $2}'").Output()
	stats.CPU.Usage, _ = strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	
	out, _ = exec.Command("sh", "-c", "cat /proc/cpuinfo | grep 'model name' | head -1 | cut -d: -f2").Output()
	stats.CPU.Model = strings.TrimSpace(string(out))
	
	// Memory
	out, _ = exec.Command("free", "-b").Output()
	lines := strings.Split(string(out), "\n")
	if len(lines) > 1 {
		fields := strings.Fields(lines[1])
		if len(fields) >= 3 {
			stats.Memory.Total, _ = strconv.ParseInt(fields[1], 10, 64)
			stats.Memory.Used, _ = strconv.ParseInt(fields[2], 10, 64)
			stats.Memory.Free, _ = strconv.ParseInt(fields[3], 10, 64)
			if stats.Memory.Total > 0 {
				stats.Memory.UsagePercent = float64(stats.Memory.Used) / float64(stats.Memory.Total) * 100
			}
		}
	}
	
	// Disk
	out, _ = exec.Command("df", "-B1", "/").Output()
	lines = strings.Split(string(out), "\n")
	if len(lines) > 1 {
		fields := strings.Fields(lines[1])
		if len(fields) >= 4 {
			stats.Disk.Total, _ = strconv.ParseInt(fields[1], 10, 64)
			stats.Disk.Used, _ = strconv.ParseInt(fields[2], 10, 64)
			stats.Disk.Free, _ = strconv.ParseInt(fields[3], 10, 64)
			if stats.Disk.Total > 0 {
				stats.Disk.UsagePercent = float64(stats.Disk.Used) / float64(stats.Disk.Total) * 100
			}
		}
	}
	
	// Load average
	out, _ = os.ReadFile("/proc/loadavg")
	fields := strings.Fields(string(out))
	if len(fields) >= 3 {
		stats.Load.Load1, _ = strconv.ParseFloat(fields[0], 64)
		stats.Load.Load5, _ = strconv.ParseFloat(fields[1], 64)
		stats.Load.Load15, _ = strconv.ParseFloat(fields[2], 64)
	}
	
	// Process count
	out, _ = exec.Command("sh", "-c", "ps aux | wc -l").Output()
	stats.Processes, _ = strconv.Atoi(strings.TrimSpace(string(out)))
	
	return stats, nil
}

// ServiceStatus contains service status
type ServiceStatus struct {
	Name    string
	Active  bool
	Enabled bool
	Memory  string
	CPU     string
}

// GetServiceStatus returns status of IronStack services
func (s *Server) GetServiceStatus() []ServiceStatus {
	services := []string{"caddy", "varnish", "mariadb", "fail2ban"}
	var statuses []ServiceStatus
	
	for _, svc := range services {
		status := ServiceStatus{Name: svc}
		
		// Check if active
		err := exec.Command("systemctl", "is-active", "--quiet", svc).Run()
		status.Active = err == nil
		
		// Check if enabled
		err = exec.Command("systemctl", "is-enabled", "--quiet", svc).Run()
		status.Enabled = err == nil
		
		// Get memory usage
		out, _ := exec.Command("sh", "-c", 
			fmt.Sprintf("systemctl show %s --property=MemoryCurrent | cut -d= -f2", svc)).Output()
		mem, _ := strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64)
		status.Memory = formatBytes(mem)
		
		statuses = append(statuses, status)
	}
	
	// Check Docker (DragonflyDB)
	dragonStatus := ServiceStatus{Name: "dragonfly"}
	out, _ := exec.Command("docker", "ps", "--filter", "name=dragonfly", "-q").Output()
	dragonStatus.Active = len(out) > 0
	statuses = append(statuses, dragonStatus)
	
	return statuses
}

// Alert represents a monitoring alert
type Alert struct {
	Level   string // "warning", "critical"
	Service string
	Message string
	Time    time.Time
}

// CheckAlerts checks for any alert conditions
func (s *Server) CheckAlerts(stats *Stats) []Alert {
	var alerts []Alert
	
	// High CPU usage
	if stats.CPU.Usage > 90 {
		alerts = append(alerts, Alert{
			Level:   "critical",
			Service: "CPU",
			Message: fmt.Sprintf("CPU usage at %.1f%%", stats.CPU.Usage),
			Time:    time.Now(),
		})
	} else if stats.CPU.Usage > 70 {
		alerts = append(alerts, Alert{
			Level:   "warning",
			Service: "CPU",
			Message: fmt.Sprintf("CPU usage at %.1f%%", stats.CPU.Usage),
			Time:    time.Now(),
		})
	}
	
	// High memory usage
	if stats.Memory.UsagePercent > 90 {
		alerts = append(alerts, Alert{
			Level:   "critical",
			Service: "Memory",
			Message: fmt.Sprintf("Memory usage at %.1f%%", stats.Memory.UsagePercent),
			Time:    time.Now(),
		})
	}
	
	// High disk usage
	if stats.Disk.UsagePercent > 90 {
		alerts = append(alerts, Alert{
			Level:   "critical",
			Service: "Disk",
			Message: fmt.Sprintf("Disk usage at %.1f%%", stats.Disk.UsagePercent),
			Time:    time.Now(),
		})
	}
	
	// High load
	if stats.Load.Load1 > float64(stats.CPU.Cores)*2 {
		alerts = append(alerts, Alert{
			Level:   "warning",
			Service: "Load",
			Message: fmt.Sprintf("Load average %.2f exceeds recommended", stats.Load.Load1),
			Time:    time.Now(),
		})
	}
	
	return alerts
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
