package crypt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeKeys создаёт пару ключей в PEM-файлах и возвращает пути к ним.
func writeKeys(t *testing.T) (privatePath, publicPath string) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	dir := t.TempDir()
	privatePath = filepath.Join(dir, "private.pem")
	publicPath = filepath.Join(dir, "public.pem")

	private := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	require.NoError(t, os.WriteFile(privatePath, private, 0600))

	public := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&key.PublicKey),
	})
	require.NoError(t, os.WriteFile(publicPath, public, 0644))

	return privatePath, publicPath
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	privatePath, publicPath := writeKeys(t)

	public, err := LoadPublicKey(publicPath)
	require.NoError(t, err)

	private, err := LoadPrivateKey(privatePath)
	require.NoError(t, err)

	message := []byte(`[{"id":"Alloc","type":"gauge","value":1.5}]`)

	encrypted, err := Encrypt(public, message)
	require.NoError(t, err)
	assert.NotEqual(t, message, encrypted)

	decrypted, err := Decrypt(private, encrypted)
	require.NoError(t, err)
	assert.Equal(t, message, decrypted)
}

func TestEncryptSplitsLongMessages(t *testing.T) {
	privatePath, publicPath := writeKeys(t)

	public, err := LoadPublicKey(publicPath)
	require.NoError(t, err)

	private, err := LoadPrivateKey(privatePath)
	require.NoError(t, err)

	message := []byte(strings.Repeat("метрика", 3000))

	encrypted, err := Encrypt(public, message)
	require.NoError(t, err)
	assert.Greater(t, len(encrypted), public.Size(), "длинное сообщение шифруется несколькими блоками")
	assert.Zero(t, len(encrypted)%public.Size())

	decrypted, err := Decrypt(private, encrypted)
	require.NoError(t, err)
	assert.Equal(t, message, decrypted)
}

func TestDecryptRejectsForeignKey(t *testing.T) {
	_, publicPath := writeKeys(t)
	otherPrivatePath, _ := writeKeys(t)

	public, err := LoadPublicKey(publicPath)
	require.NoError(t, err)

	other, err := LoadPrivateKey(otherPrivatePath)
	require.NoError(t, err)

	encrypted, err := Encrypt(public, []byte("секрет"))
	require.NoError(t, err)

	_, err = Decrypt(other, encrypted)
	assert.Error(t, err, "чужой приватный ключ не расшифровывает сообщение")
}

func TestDecryptRejectsBrokenLength(t *testing.T) {
	privatePath, _ := writeKeys(t)

	private, err := LoadPrivateKey(privatePath)
	require.NoError(t, err)

	_, err = Decrypt(private, []byte("не блок"))
	assert.Error(t, err)
}

func TestLoadKeysRejectsMissingFile(t *testing.T) {
	_, err := LoadPublicKey(filepath.Join(t.TempDir(), "missing.pem"))
	assert.Error(t, err)

	_, err = LoadPrivateKey(filepath.Join(t.TempDir(), "missing.pem"))
	assert.Error(t, err)
}

func TestLoadKeysRejectsNonPEM(t *testing.T) {
	path := filepath.Join(t.TempDir(), "broken.pem")
	require.NoError(t, os.WriteFile(path, []byte("не PEM"), 0600))

	_, err := LoadPublicKey(path)
	assert.Error(t, err)

	_, err = LoadPrivateKey(path)
	assert.Error(t, err)
}

func TestLoadPublicKeyAcceptsPKIX(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	der, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	require.NoError(t, err)

	path := filepath.Join(t.TempDir(), "public.pem")
	data := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})
	require.NoError(t, os.WriteFile(path, data, 0644))

	loaded, err := LoadPublicKey(path)
	require.NoError(t, err)
	assert.Equal(t, key.PublicKey.N, loaded.N)
}
