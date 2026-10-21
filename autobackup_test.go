package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jscyril/tailorcv/internal/domain"
)

func TestAutomaticBackupsDailyRetentionRestoreAndUndo(t *testing.T) {
	app := newWorkflowTestApp(t)
	directory := app.automaticBackupDirectory
	saveName := func(name string) {
		t.Helper()
		if _, err := app.SaveProfile(domain.ProfileInput{Name: name, Skills: []string{"Go"}}); err != nil {
			t.Fatalf("SaveProfile(%q) error = %v", name, err)
		}
	}
	profileName := func() string {
		t.Helper()
		profile, err := app.GetProfile()
		if err != nil {
			t.Fatalf("GetProfile() error = %v", err)
		}
		return profile.Name
	}

	saveName("Day One")
	day := time.Date(2026, 10, 1, 9, 0, 0, 0, time.Local)
	if err := app.writeDailyBackup(day); err != nil {
		t.Fatalf("writeDailyBackup() error = %v", err)
	}
	// A second run the same day keeps the first snapshot.
	saveName("Later Same Day")
	if err := app.writeDailyBackup(day.Add(5 * time.Hour)); err != nil {
		t.Fatalf("writeDailyBackup(same day) error = %v", err)
	}
	for offset := 1; offset <= 8; offset++ {
		if err := app.writeDailyBackup(day.AddDate(0, 0, offset)); err != nil {
			t.Fatalf("writeDailyBackup(+%d) error = %v", offset, err)
		}
	}
	backups, err := app.ListAutomaticBackups()
	if err != nil {
		t.Fatalf("ListAutomaticBackups() error = %v", err)
	}
	if len(backups) != dailyBackupsKept {
		t.Fatalf("kept %d daily backups, want %d: %#v", len(backups), dailyBackupsKept, backups)
	}
	if _, err := os.Stat(filepath.Join(directory, "tailorcv-daily-2026-10-01.json")); !os.IsNotExist(err) {
		t.Fatalf("oldest daily backup should be pruned, stat error = %v", err)
	}

	// Restore an older day, then undo it with the safety copy.
	if err := os.WriteFile(filepath.Join(directory, "tailorcv-daily-2026-09-30.json"), mustBackupJSON(t, app, "September Me"), 0o600); err != nil {
		t.Fatalf("seed old backup: %v", err)
	}
	saveName("Current Me")
	if _, err := app.RestoreAutomaticBackup("tailorcv-daily-2026-09-30.json"); err != nil {
		t.Fatalf("RestoreAutomaticBackup() error = %v", err)
	}
	if got := profileName(); got != "September Me" {
		t.Fatalf("profile after restore = %q", got)
	}
	backups, _ = app.ListAutomaticBackups()
	var safety string
	for _, backup := range backups {
		if backup.Kind == beforeRestoreBackupKind {
			safety = backup.Name
		}
	}
	if safety == "" {
		t.Fatalf("no before-restore safety copy in %#v", backups)
	}
	if _, err := app.RestoreAutomaticBackup(safety); err != nil {
		t.Fatalf("undo restore error = %v", err)
	}
	if got := profileName(); got != "Current Me" {
		t.Fatalf("profile after undo = %q", got)
	}

	for _, name := range []string{"../tailorcv-daily-2026-10-01.json", "other.json", "tailorcv-daily-2026-10-01.json/x", ""} {
		if _, err := app.RestoreAutomaticBackup(name); err == nil {
			t.Errorf("RestoreAutomaticBackup(%q) expected an error", name)
		}
	}

	// A corrupt backup is rejected before any safety copy or data change.
	if err := os.WriteFile(filepath.Join(directory, "tailorcv-daily-2026-09-29.json"), []byte("{not json"), 0o600); err != nil {
		t.Fatalf("seed corrupt backup: %v", err)
	}
	before, _ := app.ListAutomaticBackups()
	if _, err := app.RestoreAutomaticBackup("tailorcv-daily-2026-09-29.json"); err == nil {
		t.Fatal("corrupt backup restore expected an error")
	}
	after, _ := app.ListAutomaticBackups()
	if len(after) != len(before) || profileName() != "Current Me" {
		t.Fatalf("corrupt restore changed state: %d → %d backups, profile %q", len(before), len(after), profileName())
	}
}

func TestBeforeRestoreBackupsAreCapped(t *testing.T) {
	app := newWorkflowTestApp(t)
	start := time.Date(2026, 10, 1, 9, 0, 0, 0, time.Local)
	for index := 0; index < beforeRestoreBackupsKept+3; index++ {
		if err := app.writeBeforeRestoreBackup(start.Add(time.Duration(index) * time.Minute)); err != nil {
			t.Fatalf("writeBeforeRestoreBackup() error = %v", err)
		}
	}
	backups, err := app.ListAutomaticBackups()
	if err != nil || len(backups) != beforeRestoreBackupsKept {
		t.Fatalf("before-restore backups = %#v, %v", backups, err)
	}
}

func mustBackupJSON(t *testing.T, app *App, name string) []byte {
	t.Helper()
	backup, err := app.store.CreateProfileBackup(app.appContext())
	if err != nil {
		t.Fatalf("CreateProfileBackup() error = %v", err)
	}
	backup.Profile.Name = name
	data, err := json.Marshal(backup)
	if err != nil {
		t.Fatalf("marshal backup: %v", err)
	}
	return data
}
