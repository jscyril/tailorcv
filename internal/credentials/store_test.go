package credentials

import (
	"errors"
	"strings"
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

func TestVerifyLifecycleRoundTripAndCleanup(t *testing.T) {
	store := &recordedStore{values: map[string]string{}}
	if err := verifyLifecycle(store, "verification-account", "verification-secret"); err != nil {
		t.Fatalf("verifyLifecycle() error = %v", err)
	}
	if len(store.values) != 0 || store.setCalls != 1 || store.getCalls != 2 || store.deleteCalls != 1 {
		t.Fatalf("store after verification = %#v, calls = set:%d get:%d delete:%d", store.values, store.setCalls, store.getCalls, store.deleteCalls)
	}
}

func TestVerifyLifecycleCleansUpWithoutExposingSecret(t *testing.T) {
	store := &recordedStore{values: map[string]string{}, getErr: errors.New("backend unavailable")}
	err := verifyLifecycle(store, "verification-account", "must-not-appear")
	if err == nil || !strings.Contains(err.Error(), "backend unavailable") || strings.Contains(err.Error(), "must-not-appear") {
		t.Fatalf("verifyLifecycle() error = %v", err)
	}
	if len(store.values) != 0 || store.deleteCalls != 1 {
		t.Fatalf("verification credential was not cleaned up: %#v", store.values)
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

type recordedStore struct {
	values      map[string]string
	getErr      error
	setCalls    int
	getCalls    int
	deleteCalls int
}

func (store *recordedStore) Get(service, account string) (string, error) {
	store.getCalls++
	if store.getErr != nil {
		return "", store.getErr
	}
	value, ok := store.values[service+"/"+account]
	if !ok {
		return "", ErrNotFound
	}
	return value, nil
}

func (store *recordedStore) Set(service, account, secret string) error {
	store.setCalls++
	store.values[service+"/"+account] = secret
	return nil
}

func (store *recordedStore) Delete(service, account string) error {
	store.deleteCalls++
	key := service + "/" + account
	if _, ok := store.values[key]; !ok {
		return ErrNotFound
	}
	delete(store.values, key)
	return nil
}
