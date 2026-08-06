package main

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jscyril/tailorcv/internal/credentials"
	"github.com/jscyril/tailorcv/internal/storage"
)

type memoryCredentials struct{ values map[string]string }

func (store *memoryCredentials) Get(service, account string) (string, error) {
	value, ok := store.values[service+"/"+account]
	if !ok {
		return "", credentials.ErrNotFound
	}
	return value, nil
}

func (store *memoryCredentials) Set(service, account, value string) error {
	store.values[service+"/"+account] = value
	return nil
}

func (store *memoryCredentials) Delete(service, account string) error {
	key := service + "/" + account
	if _, ok := store.values[key]; !ok {
		return credentials.ErrNotFound
	}
	delete(store.values, key)
	return nil
}

func TestGeminiCredentialLifecycleNeverReturnsSecret(t *testing.T) {
	store, err := storage.Open(filepath.Join(t.TempDir(), "credentials.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close()
	keyring := &memoryCredentials{values: map[string]string{}}
	app := &App{ctx: context.Background(), store: store, credentials: keyring}
	status, err := app.GetGeminiCredentialStatus()
	if err != nil || status.Configured {
		t.Fatalf("initial status = %#v, %v", status, err)
	}
	status, err = app.SaveGeminiAPIKey(" super-secret ")
	if err != nil || !status.Configured || status.Message == "super-secret" {
		t.Fatalf("saved status = %#v, %v", status, err)
	}
	if keyring.values[credentials.Service+"/"+credentials.GeminiAPIKey] != "super-secret" {
		t.Fatal("credential was not stored in the injected keyring")
	}
	backup, err := store.CreateProfileBackup(context.Background())
	if err != nil {
		t.Fatalf("CreateProfileBackup() error = %v", err)
	}
	data, err := json.Marshal(backup)
	if err != nil || strings.Contains(string(data), "super-secret") {
		t.Fatalf("backup contains a provider credential: %v", err)
	}
	status, err = app.DeleteGeminiAPIKey()
	if err != nil || status.Configured {
		t.Fatalf("deleted status = %#v, %v", status, err)
	}
	if _, err := keyring.Get(credentials.Service, credentials.GeminiAPIKey); !errors.Is(err, credentials.ErrNotFound) {
		t.Fatalf("credential remained after delete: %v", err)
	}
}
