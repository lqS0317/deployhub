package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
)

// CryptoService AES-256-GCM 加密服务
type CryptoService struct {
	gcm cipher.AEAD
}

// NewCryptoService 根据十六进制密钥创建加密服务
func NewCryptoService(keyHex string) (*CryptoService, error) {
	keyBytes, err := hex.DecodeString(keyHex)
	if err != nil {
		return nil, fmt.Errorf("密钥十六进制解码失败: %w", err)
	}
	if len(keyBytes) != 32 {
		return nil, fmt.Errorf("密钥长度必须为 32 字节，当前为 %d 字节", len(keyBytes))
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("创建 AES 密码块失败: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("创建 GCM 模式失败: %w", err)
	}

	return &CryptoService{gcm: gcm}, nil
}

// Encrypt 加密明文，返回 Base64 编码的密文
func (s *CryptoService) Encrypt(plaintext string) (string, error) {
	nonce := make([]byte, s.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("生成随机 nonce 失败: %w", err)
	}

	ciphertext := s.gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密 Base64 编码的密文
func (s *CryptoService) Decrypt(ciphertextBase64 string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return "", fmt.Errorf("Base64 解码失败: %w", err)
	}

	nonceSize := s.gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("密文长度不足")
	}

	nonce, ciphertextBytes := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := s.gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", fmt.Errorf("解密失败: %w", err)
	}

	return string(plaintext), nil
}
