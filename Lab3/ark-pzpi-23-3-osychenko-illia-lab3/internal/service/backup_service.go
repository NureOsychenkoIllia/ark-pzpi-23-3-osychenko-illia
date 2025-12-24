package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// BackupService інтерфейс для роботи з резервними копіями
type BackupService interface {
	CreateBackup(ctx context.Context) (*BackupInfo, error)
	ListBackups(ctx context.Context) ([]BackupInfo, error)
	RestoreBackup(ctx context.Context, backupID string) error
}

// BackupInfo інформація про резервну копію
type BackupInfo struct {
	ID        string    `json:"backup_id"`
	Filename  string    `json:"filename"`
	Size      int64     `json:"size_bytes"`
	CreatedAt time.Time `json:"created_at"`
	Status    string    `json:"status"`
}

type backupService struct {
	backupDir   string
	databaseURL string
}

// NewBackupService створює новий сервіс резервного копіювання
func NewBackupService(backupDir, databaseURL string) BackupService {
	// Створюємо директорію для бекапів якщо не існує
	os.MkdirAll(backupDir, 0755)

	return &backupService{
		backupDir:   backupDir,
		databaseURL: databaseURL,
	}
}

// CreateBackup створює резервну копію бази даних
func (s *backupService) CreateBackup(ctx context.Context) (*BackupInfo, error) {
	timestamp := time.Now().Format("20060102_150405")
	backupID := fmt.Sprintf("backup_%s", timestamp)
	filename := fmt.Sprintf("%s.sql", backupID)
	filepath := filepath.Join(s.backupDir, filename)

	// Команда pg_dump для створення бекапу
	cmd := exec.CommandContext(ctx, "pg_dump", s.databaseURL, "-f", filepath)

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to create backup: %w", err)
	}

	// Отримуємо розмір файлу
	fileInfo, err := os.Stat(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to get backup file info: %w", err)
	}

	backup := &BackupInfo{
		ID:        backupID,
		Filename:  filename,
		Size:      fileInfo.Size(),
		CreatedAt: time.Now(),
		Status:    "completed",
	}

	return backup, nil
}

// ListBackups повертає список всіх резервних копій
func (s *backupService) ListBackups(ctx context.Context) ([]BackupInfo, error) {
	files, err := os.ReadDir(s.backupDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	var backups []BackupInfo
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".sql" {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		// Витягуємо ID з імені файлу
		backupID := file.Name()[:len(file.Name())-4] // видаляємо .sql

		backup := BackupInfo{
			ID:        backupID,
			Filename:  file.Name(),
			Size:      info.Size(),
			CreatedAt: info.ModTime(),
			Status:    "completed",
		}

		backups = append(backups, backup)
	}

	return backups, nil
}

// RestoreBackup відновлює базу даних з резервної копії
func (s *backupService) RestoreBackup(ctx context.Context, backupID string) error {
	filename := fmt.Sprintf("%s.sql", backupID)
	filepath := filepath.Join(s.backupDir, filename)

	// Перевіряємо чи існує файл
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return fmt.Errorf("backup file not found: %s", backupID)
	}

	// Команда psql для відновлення з бекапу
	cmd := exec.CommandContext(ctx, "psql", s.databaseURL, "-f", filepath)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	return nil
}
