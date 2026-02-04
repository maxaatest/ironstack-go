package modules

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os/exec"
)

// MariaDB manages MariaDB database
type MariaDB struct{}

func NewMariaDB() *MariaDB {
	return &MariaDB{}
}

func (m *MariaDB) Install() error {
	return exec.Command("apt-get", "install", "-y", "mariadb-server", "mariadb-client").Run()
}

func (m *MariaDB) CreateDatabase(name string) (user, password string, err error) {
	password = generatePassword()
	user = name + "_user"

	query := fmt.Sprintf(`
		CREATE DATABASE IF NOT EXISTS %s;
		CREATE USER IF NOT EXISTS '%s'@'localhost' IDENTIFIED BY '%s';
		GRANT ALL PRIVILEGES ON %s.* TO '%s'@'localhost';
		FLUSH PRIVILEGES;
	`, name, user, password, name, user)

	err = exec.Command("mysql", "-e", query).Run()
	return user, password, err
}

func (m *MariaDB) Backup(name, path string) error {
	cmd := fmt.Sprintf("mysqldump %s | gzip > %s", name, path)
	return exec.Command("sh", "-c", cmd).Run()
}

func (m *MariaDB) Status() (string, error) {
	out, err := exec.Command("systemctl", "is-active", "mariadb").Output()
	return string(out), err
}

func generatePassword() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)[:16]
}
