package credentials

import (
	"errors"
	"testing"

	keyring "github.com/zalando/go-keyring"
)

func TestOSStoreLifecycleAndNotFoundMapping(t *testing.T) {
	keyring.MockInit()
	store := OSStore{}
	if err := store.Set(Service, GeminiAPIKey, "test-secret"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	secret, err := store.Get(Service, GeminiAPIKey)
	if err != nil || secret != "test-secret" {
		t.Fatalf("Get() = %q, %v", secret, err)
	}
	if err := store.Delete(Service, GeminiAPIKey); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := store.Get(Service, GeminiAPIKey); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get() after delete error = %v", err)
	}
	if err := store.Delete(Service, GeminiAPIKey); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Delete() missing error = %v", err)
	}
}

func TestOSStorePreservesBackendErrors(t *testing.T) {
	backendErr := errors.New("credential backend unavailable")
	keyring.MockInitWithError(backendErr)
	store := OSStore{}
	if _, err := store.Get(Service, GeminiAPIKey); !errors.Is(err, backendErr) {
		t.Fatalf("Get() error = %v", err)
	}
	if err := store.Set(Service, GeminiAPIKey, "test-secret"); !errors.Is(err, backendErr) {
		t.Fatalf("Set() error = %v", err)
	}
	if err := store.Delete(Service, GeminiAPIKey); !errors.Is(err, backendErr) {
		t.Fatalf("Delete() error = %v", err)
	}
}
