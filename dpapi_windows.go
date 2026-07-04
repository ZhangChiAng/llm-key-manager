//go:build windows

package main

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

func protectKeyValue(value string) ([]byte, error) {
	input := []byte(value)
	dataIn := dataBlob(input)
	var dataOut windows.DataBlob

	if err := windows.CryptProtectData(&dataIn, nil, nil, 0, nil, windows.CRYPTPROTECT_UI_FORBIDDEN, &dataOut); err != nil {
		return nil, err
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(dataOut.Data)))

	return dataBlobBytes(dataOut), nil
}

func unprotectKeyValue(value []byte) (string, error) {
	dataIn := dataBlob(value)
	var description *uint16
	var dataOut windows.DataBlob

	if err := windows.CryptUnprotectData(&dataIn, &description, nil, 0, nil, windows.CRYPTPROTECT_UI_FORBIDDEN, &dataOut); err != nil {
		return "", err
	}
	if description != nil {
		defer windows.LocalFree(windows.Handle(unsafe.Pointer(description)))
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(dataOut.Data)))

	return string(dataBlobBytes(dataOut)), nil
}

func dataBlob(value []byte) windows.DataBlob {
	if len(value) == 0 {
		return windows.DataBlob{}
	}

	return windows.DataBlob{
		Size: uint32(len(value)),
		Data: &value[0],
	}
}

func dataBlobBytes(blob windows.DataBlob) []byte {
	if blob.Size == 0 {
		return []byte{}
	}

	value := make([]byte, blob.Size)
	copy(value, unsafe.Slice(blob.Data, blob.Size))

	return value
}
