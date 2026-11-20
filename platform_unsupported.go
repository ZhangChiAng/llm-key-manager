//go:build !windows && !darwin

package main

import "errors"

// This stub only lets the package type-check and lint on the Linux development host.
var errKeyStorageUnsupported = errors.New("key storage is only supported on Windows and macOS")

func protectKeyValue(string) ([]byte, error) {
	return nil, errKeyStorageUnsupported
}

func unprotectKeyValue([]byte) (string, error) {
	return "", errKeyStorageUnsupported
}

func keyStoreBaseDir() (string, error) {
	return "", errKeyStorageUnsupported
}
