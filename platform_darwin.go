//go:build darwin

package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// The AES-256-GCM master key is a generic password in the user's login Keychain,
// accessed through Apple's security tool by absolute path (no cgo, no PATH lookup).
// Items created by that tool list it in their ACL, so rebuilding the unsigned app does
// not trigger a Keychain prompt on every read. Like DPAPI without extra entropy, this
// binds decryption to the logged-in OS user rather than to this binary.
const (
	securityToolPath     = "/usr/bin/security"
	keychainService      = "llm-key-manager"
	keychainAccount      = "master-key"
	masterKeySize        = 32
	securityItemNotFound = 44
)

var (
	// masterKeyMu guards masterKey. ProbeKey runs concurrently and ListKeys decrypts
	// every record, so the key is read or created once per process, not per record.
	masterKeyMu sync.Mutex
	masterKey   []byte

	errMasterKeyNotFound = errors.New("keychain master key not found")
)

// keyStoreBaseDir returns ~/Library/Application Support.
func keyStoreBaseDir() (string, error) {
	return os.UserConfigDir()
}

// protectKeyValue returns nonce || ciphertext+tag and creates the master key on first use.
func protectKeyValue(value string) ([]byte, error) {
	gcm, err := masterKeyCipher(true)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, []byte(value), nil), nil
}

// unprotectKeyValue never creates a master key: a new key cannot decrypt existing records
// and would hide that the original Keychain item was lost.
func unprotectKeyValue(value []byte) (string, error) {
	gcm, err := masterKeyCipher(false)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(value) < nonceSize+gcm.Overhead() {
		return "", errors.New("encrypted key value is too short")
	}

	plaintext, err := gcm.Open(nil, value[:nonceSize], value[nonceSize:], nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func masterKeyCipher(create bool) (cipher.AEAD, error) {
	key, err := loadMasterKey(create)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	return cipher.NewGCM(block)
}

func loadMasterKey(create bool) ([]byte, error) {
	masterKeyMu.Lock()
	defer masterKeyMu.Unlock()

	if masterKey != nil {
		return masterKey, nil
	}

	key, err := readKeychainMasterKey()
	if errors.Is(err, errMasterKeyNotFound) {
		if !create {
			return nil, fmt.Errorf("%w (service %q, account %q); saved keys cannot be decrypted", errMasterKeyNotFound, keychainService, keychainAccount)
		}
		key, err = createKeychainMasterKey()
	}
	if err != nil {
		return nil, err
	}

	masterKey = key
	return masterKey, nil
}

func readKeychainMasterKey() ([]byte, error) {
	output, err := exec.Command(securityToolPath, "find-generic-password", "-s", keychainService, "-a", keychainAccount, "-w").Output()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		if exitErr.ExitCode() == securityItemNotFound {
			return nil, errMasterKeyNotFound
		}
		return nil, fmt.Errorf("read keychain master key: %w: %s", err, strings.TrimSpace(string(exitErr.Stderr)))
	}
	if err != nil {
		return nil, fmt.Errorf("read keychain master key: %w", err)
	}

	key, err := hex.DecodeString(strings.TrimSpace(string(output)))
	if err != nil || len(key) != masterKeySize {
		return nil, errors.New("keychain master key is invalid")
	}

	return key, nil
}

func createKeychainMasterKey() ([]byte, error) {
	key := make([]byte, masterKeySize)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}

	// The secret goes to `security -i` on stdin so it never appears in argv. Without -U,
	// add fails instead of overwriting an existing item. Interactive mode's output may
	// contain the input line and is discarded, and its exit status does not reliably
	// reflect command failure, so the read-back below is the authoritative check.
	command := fmt.Sprintf("add-generic-password -s %s -a %s -w %s\n", keychainService, keychainAccount, hex.EncodeToString(key))
	cmd := exec.Command(securityToolPath, "-i")
	cmd.Stdin = strings.NewReader(command)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("create keychain master key: %w", err)
	}

	storedKey, err := readKeychainMasterKey()
	if err != nil {
		return nil, fmt.Errorf("verify keychain master key: %w", err)
	}
	if !bytes.Equal(storedKey, key) {
		return nil, errors.New("keychain master key does not match the generated key")
	}

	return key, nil
}
