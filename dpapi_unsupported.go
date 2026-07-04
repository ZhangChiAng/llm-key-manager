//go:build !windows

package main

import "errors"

func protectKeyValue(string) ([]byte, error) {
	return nil, errors.New("DPAPI key storage is only available on Windows")
}

func unprotectKeyValue([]byte) (string, error) {
	return "", errors.New("DPAPI key storage is only available on Windows")
}
