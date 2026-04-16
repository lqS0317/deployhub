package crypto

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	// 32 字节密钥（64 位十六进制）
	keyHex := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	svc, err := NewCryptoService(keyHex)
	require.NoError(t, err)

	plaintext := "这是一段敏感数据：kubeconfig内容"
	encrypted, err := svc.Encrypt(plaintext)
	require.NoError(t, err)
	assert.NotEqual(t, plaintext, encrypted)

	decrypted, err := svc.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncryptDecryptEmptyString(t *testing.T) {
	keyHex := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	svc, err := NewCryptoService(keyHex)
	require.NoError(t, err)

	encrypted, err := svc.Encrypt("")
	require.NoError(t, err)

	decrypted, err := svc.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, "", decrypted)
}

func TestInvalidKeyLength(t *testing.T) {
	_, err := NewCryptoService("tooshort")
	assert.Error(t, err)
}

func TestDecryptWithWrongKey(t *testing.T) {
	key1 := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	key2 := "fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210"

	svc1, _ := NewCryptoService(key1)
	svc2, _ := NewCryptoService(key2)

	encrypted, err := svc1.Encrypt("secret data")
	require.NoError(t, err)

	_, err = svc2.Decrypt(encrypted)
	assert.Error(t, err)
}

func TestEncryptProducesDifferentCiphertext(t *testing.T) {
	keyHex := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	svc, _ := NewCryptoService(keyHex)

	e1, _ := svc.Encrypt("same text")
	e2, _ := svc.Encrypt("same text")
	// AES-GCM 使用随机 nonce，相同明文产生不同密文
	assert.NotEqual(t, e1, e2)
}

func TestValidKeyHex(t *testing.T) {
	// 确保密钥正确解析
	keyHex := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	keyBytes, err := hex.DecodeString(keyHex)
	require.NoError(t, err)
	assert.Equal(t, 32, len(keyBytes))
}
