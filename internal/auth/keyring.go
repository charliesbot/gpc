package auth

import (
	"fmt"

	"github.com/99designs/keyring"
)

const (
	serviceName = "gpc"
	keyName     = "service-account-key"
)

// allowedBackends restricts keyring storage to OS-native credential stores.
var allowedBackends = []keyring.BackendType{
	keyring.KeychainBackend,       // macOS Keychain
	keyring.SecretServiceBackend,  // Linux (D-Bus Secret Service / GNOME Keyring)
	keyring.WinCredBackend,        // Windows Credential Manager
}

// openKeyring returns a configured keyring scoped to the gpc service.
func openKeyring() (keyring.Keyring, error) {
	ring, err := keyring.Open(keyring.Config{
		ServiceName:     serviceName,
		AllowedBackends: allowedBackends,

		// macOS Keychain settings
		KeychainName:                   "login",
		KeychainTrustApplication:       true,
		KeychainSynchronizable:         false,
		KeychainAccessibleWhenUnlocked: true,
	})
	if err != nil {
		return nil, fmt.Errorf("opening system keyring: %w", err)
	}
	return ring, nil
}

// StoreCredential saves a service account JSON key to the system keychain.
func StoreCredential(jsonKey []byte) error {
	ring, err := openKeyring()
	if err != nil {
		return err
	}

	err = ring.Set(keyring.Item{
		Key:         keyName,
		Data:        jsonKey,
		Label:       "GPC Service Account Key",
		Description: "Google Play Console service account credentials",
	})
	if err != nil {
		return fmt.Errorf("storing credential in keyring: %w", err)
	}
	return nil
}

// LoadCredential retrieves the service account JSON key from the system keychain.
// Returns keyring.ErrKeyNotFound if no credential has been stored.
func LoadCredential() ([]byte, error) {
	ring, err := openKeyring()
	if err != nil {
		return nil, err
	}

	item, err := ring.Get(keyName)
	if err != nil {
		return nil, fmt.Errorf("loading credential from keyring: %w", err)
	}
	return item.Data, nil
}

// DeleteCredential removes the service account JSON key from the system keychain.
func DeleteCredential() error {
	ring, err := openKeyring()
	if err != nil {
		return err
	}

	err = ring.Remove(keyName)
	if err != nil {
		return fmt.Errorf("deleting credential from keyring: %w", err)
	}
	return nil
}
