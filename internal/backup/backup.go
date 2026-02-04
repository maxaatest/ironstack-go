package backup

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Manager handles backup operations
type Manager struct {
	BackupDir string
}

// New creates a backup manager
func New() *Manager {
	return &Manager{BackupDir: "/backups"}
}

// Backup represents a backup file
type Backup struct {
	Name      string
	Path      string
	Size      int64
	Created   time.Time
	Type      string // "full", "db", "files"
}

// CreateFull creates a full backup (files + database)
func (m *Manager) CreateFull(sitePath, domain string) (*Backup, error) {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	backupName := fmt.Sprintf("%s_full_%s", domain, timestamp)
	backupPath := filepath.Join(m.BackupDir, domain, backupName+".tar.gz")

	// Create backup directory
	os.MkdirAll(filepath.Dir(backupPath), 0755)

	// First, export the database
	dbBackup := filepath.Join(sitePath, "database.sql")
	if err := exec.Command("wp", "db", "export", dbBackup, "--path="+sitePath+"/public").Run(); err != nil {
		return nil, fmt.Errorf("database export failed: %w", err)
	}

	// Create tar.gz of entire site
	cmd := exec.Command("tar", "-czf", backupPath, "-C", filepath.Dir(sitePath), filepath.Base(sitePath))
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("tar failed: %w", err)
	}

	// Clean up SQL file
	os.Remove(dbBackup)

	// Get file info
	info, _ := os.Stat(backupPath)

	return &Backup{
		Name:    backupName,
		Path:    backupPath,
		Size:    info.Size(),
		Created: time.Now(),
		Type:    "full",
	}, nil
}

// CreateDBOnly creates a database-only backup
func (m *Manager) CreateDBOnly(sitePath, domain string) (*Backup, error) {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	backupName := fmt.Sprintf("%s_db_%s", domain, timestamp)
	backupPath := filepath.Join(m.BackupDir, domain, backupName+".sql.gz")

	// Create backup directory
	os.MkdirAll(filepath.Dir(backupPath), 0755)

	// Export database
	sqlPath := filepath.Join(m.BackupDir, domain, backupName+".sql")
	if err := exec.Command("wp", "db", "export", sqlPath, "--path="+sitePath+"/public").Run(); err != nil {
		return nil, fmt.Errorf("database export failed: %w", err)
	}

	// Compress
	if err := gzipFile(sqlPath, backupPath); err != nil {
		return nil, fmt.Errorf("compression failed: %w", err)
	}

	// Clean up uncompressed
	os.Remove(sqlPath)

	info, _ := os.Stat(backupPath)

	return &Backup{
		Name:    backupName,
		Path:    backupPath,
		Size:    info.Size(),
		Created: time.Now(),
		Type:    "db",
	}, nil
}

// CreateFilesOnly creates a files-only backup
func (m *Manager) CreateFilesOnly(sitePath, domain string) (*Backup, error) {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	backupName := fmt.Sprintf("%s_files_%s", domain, timestamp)
	backupPath := filepath.Join(m.BackupDir, domain, backupName+".tar.gz")

	// Create backup directory
	os.MkdirAll(filepath.Dir(backupPath), 0755)

	// Create tar.gz excluding uploads (large files)
	cmd := exec.Command("tar", "-czf", backupPath,
		"--exclude=wp-content/uploads",
		"-C", filepath.Dir(sitePath), filepath.Base(sitePath))
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("tar failed: %w", err)
	}

	info, _ := os.Stat(backupPath)

	return &Backup{
		Name:    backupName,
		Path:    backupPath,
		Size:    info.Size(),
		Created: time.Now(),
		Type:    "files",
	}, nil
}

// Restore restores a backup
func (m *Manager) Restore(backupPath, sitePath string) error {
	switch {
	case filepath.Ext(backupPath) == ".gz" && filepath.Ext(backupPath[:len(backupPath)-3]) == ".sql":
		// Database-only restore
		sqlPath := backupPath[:len(backupPath)-3]
		if err := gunzipFile(backupPath, sqlPath); err != nil {
			return fmt.Errorf("decompress failed: %w", err)
		}
		if err := exec.Command("wp", "db", "import", sqlPath, "--path="+sitePath+"/public").Run(); err != nil {
			return fmt.Errorf("import failed: %w", err)
		}
		os.Remove(sqlPath)
	default:
		// Full restore
		if err := exec.Command("tar", "-xzf", backupPath, "-C", filepath.Dir(sitePath)).Run(); err != nil {
			return fmt.Errorf("extract failed: %w", err)
		}
	}

	return nil
}

// List returns all backups for a domain
func (m *Manager) List(domain string) ([]Backup, error) {
	backupDir := filepath.Join(m.BackupDir, domain)
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return nil, err
	}

	var backups []Backup
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, _ := entry.Info()
		backupType := "full"
		if filepath.Ext(entry.Name()) == ".gz" {
			name := entry.Name()
			if len(name) > 7 && name[len(name)-7:] == ".sql.gz" {
				backupType = "db"
			}
		}
		if ok, _ := filepath.Match("*_files_*", entry.Name()); ok {
			backupType = "files"
		}

		backups = append(backups, Backup{
			Name:    entry.Name(),
			Path:    filepath.Join(backupDir, entry.Name()),
			Size:    info.Size(),
			Created: info.ModTime(),
			Type:    backupType,
		})
	}

	return backups, nil
}

// Delete removes a backup
func (m *Manager) Delete(backupPath string) error {
	return os.Remove(backupPath)
}

// helpers

func gzipFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	gw := gzip.NewWriter(out)
	defer gw.Close()

	_, err = io.Copy(gw, in)
	return err
}

func gunzipFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	gr, err := gzip.NewReader(in)
	if err != nil {
		return err
	}
	defer gr.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, gr)
	return err
}
