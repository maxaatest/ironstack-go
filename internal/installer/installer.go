package installer

import (
	"fmt"
	"os/exec"
	"runtime"
)

// Component represents an installable component
type Component struct {
	Name    string
	Install func() error
	Check   func() bool
}

// Installer manages component installation
type Installer struct {
	components []Component
}

// New creates a new installer with all components
func New() *Installer {
	return &Installer{
		components: []Component{
			{Name: "Caddy", Install: installCaddy, Check: checkCaddy},
			{Name: "Varnish", Install: installVarnish, Check: checkVarnish},
			{Name: "MariaDB", Install: installMariaDB, Check: checkMariaDB},
			{Name: "DragonflyDB", Install: installDragonfly, Check: checkDragonfly},
			{Name: "WP-CLI", Install: installWPCLI, Check: checkWPCLI},
			{Name: "CSF", Install: installCSF, Check: checkCSF},
			{Name: "Fail2ban", Install: installFail2ban, Check: checkFail2ban},
			{Name: "GoAccess", Install: installGoAccess, Check: checkGoAccess},
		},
	}
}

// Components returns all components
func (i *Installer) Components() []Component {
	return i.components
}

// InstallAll installs all components
func (i *Installer) InstallAll(progress func(name string, done bool)) error {
	for _, c := range i.components {
		progress(c.Name, false)
		if err := c.Install(); err != nil {
			return fmt.Errorf("failed to install %s: %w", c.Name, err)
		}
		progress(c.Name, true)
	}
	return nil
}

// CheckRequirements verifies system requirements
func CheckRequirements() error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("IronStack requires Linux, detected: %s", runtime.GOOS)
	}
	
	// Check if running as root
	if err := exec.Command("id", "-u").Run(); err != nil {
		return fmt.Errorf("failed to check user: %w", err)
	}
	
	return nil
}

// --- Caddy ---
func installCaddy() error {
	commands := []string{
		"apt-get update",
		"apt-get install -y debian-keyring debian-archive-keyring apt-transport-https curl",
		"curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg",
		"curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | tee /etc/apt/sources.list.d/caddy-stable.list",
		"apt-get update",
		"apt-get install -y caddy",
	}
	return runCommands(commands)
}

func checkCaddy() bool {
	return commandExists("caddy")
}

// --- Varnish ---
func installVarnish() error {
	commands := []string{
		"apt-get install -y varnish",
		"systemctl enable varnish",
	}
	return runCommands(commands)
}

func checkVarnish() bool {
	return commandExists("varnishd")
}

// --- MariaDB ---
func installMariaDB() error {
	commands := []string{
		"apt-get install -y mariadb-server mariadb-client",
		"systemctl enable mariadb",
		"systemctl start mariadb",
	}
	return runCommands(commands)
}

func checkMariaDB() bool {
	return commandExists("mysql")
}

// --- DragonflyDB ---
func installDragonfly() error {
	commands := []string{
		"curl -fsSL https://get.docker.com | sh",
		"docker pull docker.dragonflydb.io/dragonflydb/dragonfly",
		"docker run -d --name dragonfly --restart=always -p 6379:6379 docker.dragonflydb.io/dragonflydb/dragonfly",
	}
	return runCommands(commands)
}

func checkDragonfly() bool {
	out, _ := exec.Command("docker", "ps", "--filter", "name=dragonfly", "-q").Output()
	return len(out) > 0
}

// --- WP-CLI ---
func installWPCLI() error {
	commands := []string{
		"curl -O https://raw.githubusercontent.com/wp-cli/builds/gh-pages/phar/wp-cli.phar",
		"chmod +x wp-cli.phar",
		"mv wp-cli.phar /usr/local/bin/wp",
	}
	return runCommands(commands)
}

func checkWPCLI() bool {
	return commandExists("wp")
}

// --- CSF ---
func installCSF() error {
	commands := []string{
		"cd /usr/src && curl -O https://download.configserver.com/csf.tgz",
		"cd /usr/src && tar -xzf csf.tgz",
		"cd /usr/src/csf && sh install.sh",
	}
	return runCommands(commands)
}

func checkCSF() bool {
	return commandExists("csf")
}

// --- Fail2ban ---
func installFail2ban() error {
	return runCommands([]string{"apt-get install -y fail2ban", "systemctl enable fail2ban"})
}

func checkFail2ban() bool {
	return commandExists("fail2ban-client")
}

// --- GoAccess ---
func installGoAccess() error {
	return runCommands([]string{"apt-get install -y goaccess"})
}

func checkGoAccess() bool {
	return commandExists("goaccess")
}

// --- Helpers ---
func runCommands(commands []string) error {
	for _, cmd := range commands {
		if err := exec.Command("sh", "-c", cmd).Run(); err != nil {
			return fmt.Errorf("command failed: %s: %w", cmd, err)
		}
	}
	return nil
}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
