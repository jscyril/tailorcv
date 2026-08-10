package credentials

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"

	keyring "github.com/zalando/go-keyring"
)

const (
	Service             = "TailorCV"
	GeminiAPIKey        = "gemini-api-key"
	verificationService = "TailorCV Native Verification"
)

var ErrNotFound = errors.New("credential not found")

// Store keeps secrets outside TailorCV's database and backups.
type Store interface {
	Get(service, account string) (string, error)
	Set(service, account, secret string) error
	Delete(service, account string) error
}

type OSStore struct{}

func (OSStore) Get(service, account string) (string, error) {
	secret, err := keyring.Get(service, account)
	if errors.Is(err, keyring.ErrNotFound) {
		return "", ErrNotFound
	}
	return secret, err
}

func (OSStore) Set(service, account, secret string) error {
	return keyring.Set(service, account, secret)
}

func (OSStore) Delete(service, account string) error {
	err := keyring.Delete(service, account)
	if errors.Is(err, keyring.ErrNotFound) {
		return ErrNotFound
	}
	return err
}

// VerifyNativeLifecycle performs a disposable set/get/delete check against the
// operating-system credential backend. Generated identifiers and secrets are
// never returned or included in errors, and cleanup is attempted on every exit.
func VerifyNativeLifecycle() error {
	accountBytes := make([]byte, 12)
	secretBytes := make([]byte, 32)
	if _, err := rand.Read(accountBytes); err != nil {
		return fmt.Errorf("generate verification account: %w", err)
	}
	if _, err := rand.Read(secretBytes); err != nil {
		return fmt.Errorf("generate verification secret: %w", err)
	}
	return verifyLifecycle(OSStore{}, "tailorcv-verification-"+hex.EncodeToString(accountBytes), hex.EncodeToString(secretBytes))
}

func verifyLifecycle(store Store, account, secret string) error {
	cleanupRequired := false
	defer func() {
		if cleanupRequired {
			_ = store.Delete(verificationService, account)
		}
	}()

	if err := store.Set(verificationService, account, secret); err != nil {
		return fmt.Errorf("set native verification credential: %w", err)
	}
	cleanupRequired = true
	stored, err := store.Get(verificationService, account)
	if err != nil {
		return fmt.Errorf("get native verification credential: %w", err)
	}
	if subtle.ConstantTimeCompare([]byte(stored), []byte(secret)) != 1 {
		return fmt.Errorf("native verification credential did not round-trip")
	}
	if err := store.Delete(verificationService, account); err != nil {
		return fmt.Errorf("delete native verification credential: %w", err)
	}
	cleanupRequired = false
	if _, err := store.Get(verificationService, account); !errors.Is(err, ErrNotFound) {
		if err == nil {
			return fmt.Errorf("native verification credential remained after deletion")
		}
		return fmt.Errorf("confirm native verification credential deletion: %w", err)
	}
	return nil
}
