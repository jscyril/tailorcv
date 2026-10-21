package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	backupfile "github.com/jscyril/tailorcv/internal/backup"
	"github.com/jscyril/tailorcv/internal/domain"
)

const (
	dailyBackupKind          = "daily"
	beforeRestoreBackupKind  = "before-restore"
	dailyBackupsKept         = 7
	beforeRestoreBackupsKept = 5
)

// Names look like tailorcv-daily-2026-10-21.json or
// tailorcv-before-restore-20261021-153000.json.
var automaticBackupName = regexp.MustCompile(`^tailorcv-(daily|before-restore)-[0-9-]+\.json$`)

func (a *App) backupDirectory() (string, error) {
	directory := a.automaticBackupDirectory
	if directory == "" {
		configDir, err := os.UserConfigDir()
		if err != nil {
			return "", fmt.Errorf("resolve backup directory: %w", err)
		}
		directory = filepath.Join(configDir, "tailorcv", "backups")
	}
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return "", fmt.Errorf("create backup directory: %w", err)
	}
	return directory, nil
}

// runAutomaticBackups writes today's backup now and re-checks hourly, so an
// app left open for days still gets one backup per day.
func (a *App) runAutomaticBackups(ctx context.Context) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		if err := a.writeDailyBackup(time.Now()); err != nil {
			fmt.Printf("automatic backup: %v\n", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// writeDailyBackup writes one backup per local calendar day and keeps the newest few.
func (a *App) writeDailyBackup(now time.Time) error {
	a.backupMu.Lock()
	defer a.backupMu.Unlock()
	if a.store == nil {
		return errors.New("local profile store is not open")
	}
	directory, err := a.backupDirectory()
	if err != nil {
		return err
	}
	path := filepath.Join(directory, "tailorcv-daily-"+now.Format("2006-01-02")+".json")
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("check daily backup: %w", err)
	}
	if err := a.writeBackupFile(path); err != nil {
		return err
	}
	return pruneAutomaticBackups(directory, dailyBackupKind, dailyBackupsKept)
}

// writeBeforeRestoreBackup saves current data so a wrong restore can be undone.
func (a *App) writeBeforeRestoreBackup(now time.Time) error {
	a.backupMu.Lock()
	defer a.backupMu.Unlock()
	directory, err := a.backupDirectory()
	if err != nil {
		return err
	}
	path := filepath.Join(directory, "tailorcv-before-restore-"+now.Format("20060102-150405")+".json")
	if err := a.writeBackupFile(path); err != nil {
		return fmt.Errorf("save a safety copy before restoring: %w", err)
	}
	return pruneAutomaticBackups(directory, beforeRestoreBackupKind, beforeRestoreBackupsKept)
}

func (a *App) writeBackupFile(path string) error {
	backup, err := a.store.CreateProfileBackup(a.appContext())
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(backup, "", "  ")
	if err != nil {
		return fmt.Errorf("encode backup: %w", err)
	}
	return backupfile.Write(path, append(data, '\n'))
}

func listAutomaticBackups(directory string) ([]domain.AutomaticBackup, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("list automatic backups: %w", err)
	}
	backups := make([]domain.AutomaticBackup, 0)
	for _, entry := range entries {
		match := automaticBackupName.FindStringSubmatch(entry.Name())
		if match == nil || !entry.Type().IsRegular() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		backups = append(backups, domain.AutomaticBackup{Name: entry.Name(), Kind: match[1], CreatedAt: info.ModTime().UTC().Format(time.RFC3339), Size: info.Size()})
	}
	sort.Slice(backups, func(i, j int) bool { return backups[i].CreatedAt > backups[j].CreatedAt })
	return backups, nil
}

func pruneAutomaticBackups(directory, kind string, keep int) error {
	backups, err := listAutomaticBackups(directory)
	if err != nil {
		return err
	}
	var names []string
	for _, backup := range backups {
		if backup.Kind == kind {
			names = append(names, backup.Name)
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(names)))
	for index := keep; index < len(names); index++ {
		if err := os.Remove(filepath.Join(directory, names[index])); err != nil {
			return fmt.Errorf("remove old backup: %w", err)
		}
	}
	return nil
}

// ListAutomaticBackups returns daily and before-restore backups, newest first.
func (a *App) ListAutomaticBackups() ([]domain.AutomaticBackup, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	directory, err := a.backupDirectory()
	if err != nil {
		return nil, err
	}
	return listAutomaticBackups(directory)
}

// RestoreAutomaticBackup restores one listed automatic backup by name.
func (a *App) RestoreAutomaticBackup(name string) (domain.BackupResult, error) {
	if err := a.ready(); err != nil {
		return domain.BackupResult{}, err
	}
	if !automaticBackupName.MatchString(name) || strings.ContainsAny(name, `/\`) {
		return domain.BackupResult{}, fmt.Errorf("automatic backup name is not valid")
	}
	directory, err := a.backupDirectory()
	if err != nil {
		return domain.BackupResult{}, err
	}
	path := filepath.Join(directory, name)
	data, err := backupfile.Read(path)
	if err != nil {
		return domain.BackupResult{}, err
	}
	return a.restoreBackup(data, path)
}

// restoreBackup validates data completely, saves a safety copy of current
// data, then replaces everything in one transaction.
func (a *App) restoreBackup(data []byte, path string) (domain.BackupResult, error) {
	backup, err := domain.DecodeProfileBackup(data)
	if err != nil {
		return domain.BackupResult{}, err
	}
	if err := a.writeBeforeRestoreBackup(time.Now()); err != nil {
		return domain.BackupResult{}, err
	}
	if err := a.store.ReplaceProfileFromBackup(a.appContext(), backup); err != nil {
		return domain.BackupResult{}, err
	}
	return backup.Result(path), nil
}
