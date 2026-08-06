package credentials

import (
	"errors"

	keyring "github.com/zalando/go-keyring"
)

const (
	Service      = "TailorCV"
	GeminiAPIKey = "gemini-api-key"
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
