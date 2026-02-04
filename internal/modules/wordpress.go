package modules

import (
	"os/exec"
)

// WordPress manages WordPress via WP-CLI
type WordPress struct{}

func NewWordPress() *WordPress {
	return &WordPress{}
}

func (w *WordPress) InstallCLI() error {
	cmd := `curl -O https://raw.githubusercontent.com/wp-cli/builds/gh-pages/phar/wp-cli.phar && 
		chmod +x wp-cli.phar && 
		mv wp-cli.phar /usr/local/bin/wp`
	return exec.Command("sh", "-c", cmd).Run()
}

func (w *WordPress) Install(path, url, dbName, dbUser, dbPass string) error {
	// Download WordPress
	if err := exec.Command("wp", "core", "download", "--path="+path).Run(); err != nil {
		return err
	}

	// Create config
	if err := exec.Command("wp", "config", "create",
		"--path="+path,
		"--dbname="+dbName,
		"--dbuser="+dbUser,
		"--dbpass="+dbPass,
	).Run(); err != nil {
		return err
	}

	return nil
}

func (w *WordPress) AutoTune(path string) error {
	settings := []string{
		"WP_MEMORY_LIMIT='256M'",
		"WP_MAX_MEMORY_LIMIT='512M'",
		"WP_POST_REVISIONS=5",
		"AUTOSAVE_INTERVAL=120",
		"EMPTY_TRASH_DAYS=7",
		"DISABLE_WP_CRON=true",
		"WP_CACHE=true",
	}

	for _, setting := range settings {
		exec.Command("wp", "config", "set", setting, "--path="+path, "--raw").Run()
	}

	// Install and enable Redis cache
	exec.Command("wp", "plugin", "install", "redis-cache", "--activate", "--path="+path).Run()
	exec.Command("wp", "redis", "enable", "--path="+path).Run()

	return nil
}
