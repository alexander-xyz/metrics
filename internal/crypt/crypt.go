// Package crypt шифрует сообщения от агента к серверу асимметричным
// алгоритмом RSA: агент шифрует публичным ключом, сервер расшифровывает
// приватным.
package crypt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

// ErrNotRSAKey возвращается, когда файл содержит ключ другого алгоритма.
var ErrNotRSAKey = errors.New("key is not an RSA key")

// LoadPublicKey читает публичный ключ из PEM-файла.
func LoadPublicKey(path string) (*rsa.PublicKey, error) {
	block, err := readPEM(path)
	if err != nil {
		return nil, err
	}

	if key, parseErr := x509.ParsePKCS1PublicKey(block.Bytes); parseErr == nil {
		return key, nil
	}

	parsed, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse public key %s: %w", path, err)
	}

	key, ok := parsed.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("%s: %w", path, ErrNotRSAKey)
	}

	return key, nil
}

// LoadPrivateKey читает приватный ключ из PEM-файла.
func LoadPrivateKey(path string) (*rsa.PrivateKey, error) {
	block, err := readPEM(path)
	if err != nil {
		return nil, err
	}

	if key, parseErr := x509.ParsePKCS1PrivateKey(block.Bytes); parseErr == nil {
		return key, nil
	}

	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key %s: %w", path, err)
	}

	key, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("%s: %w", path, ErrNotRSAKey)
	}

	return key, nil
}

// readPEM читает файл и выделяет из него первый PEM-блок.
func readPEM(path string) (*pem.Block, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read key file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("%s: no PEM data found", path)
	}

	return block, nil
}

// Encrypt шифрует данные публичным ключом. Данные длиннее одного блока
// шифруются по частям: размер блока ограничен длиной ключа.
func Encrypt(key *rsa.PublicKey, data []byte) ([]byte, error) {
	hash := sha256.New()
	limit := key.Size() - 2*hash.Size() - 2

	if limit <= 0 {
		return nil, fmt.Errorf("key size %d is too small", key.Size())
	}

	encrypted := make([]byte, 0, len(data)+key.Size())

	for len(data) > 0 {
		size := min(limit, len(data))

		chunk, err := rsa.EncryptOAEP(hash, rand.Reader, key, data[:size], nil)
		if err != nil {
			return nil, fmt.Errorf("encrypt chunk: %w", err)
		}

		encrypted = append(encrypted, chunk...)
		data = data[size:]
	}

	return encrypted, nil
}

// Decrypt расшифровывает данные приватным ключом.
func Decrypt(key *rsa.PrivateKey, data []byte) ([]byte, error) {
	size := key.PublicKey.Size()

	if len(data)%size != 0 {
		return nil, fmt.Errorf("encrypted data length %d is not a multiple of %d", len(data), size)
	}

	hash := sha256.New()
	decrypted := make([]byte, 0, len(data))

	for len(data) > 0 {
		chunk, err := rsa.DecryptOAEP(hash, rand.Reader, key, data[:size], nil)
		if err != nil {
			return nil, fmt.Errorf("decrypt chunk: %w", err)
		}

		decrypted = append(decrypted, chunk...)
		data = data[size:]
	}

	return decrypted, nil
}
