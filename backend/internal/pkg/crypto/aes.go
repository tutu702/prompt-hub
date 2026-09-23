// Package crypto - AES-GCM 对称加密工具（PRD P08）。
//
// 用途：ph_llm_configs.api_key_enc 加密存储。
// 密钥：环境变量 PROMPTHUB_LLM_AES_KEY（64 个 hex 字符 = 32 字节）；缺失则在 New() 报错。
// 密文布局：nonce(12B) || ciphertext_with_tag，整体再 hex。
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"os"
)

const (
	envKeyName = "PROMPTHUB_LLM_AES_KEY"
	nonceSize  = 12
)

// AES AES-GCM 加解密器。
type AES struct {
	gcm cipher.AEAD
}

// New 从环境变量加载密钥并构造 AES。
func New() (*AES, error) {
	raw := os.Getenv(envKeyName)
	if raw == "" {
		return nil, errors.New("PROMPTHUB_LLM_AES_KEY 未配置（需 32 字节 hex）")
	}
	key, err := hex.DecodeString(raw)
	if err != nil || len(key) != 32 {
		return nil, errors.New("PROMPTHUB_LLM_AES_KEY 必须为 64 个十六进制字符")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	g, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &AES{gcm: g}, nil
}

// Encrypt 明文 → hex(nonce+ciphertext)。
func (a *AES) Encrypt(plain string) (string, error) {
	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ct := a.gcm.Seal(nil, nonce, []byte(plain), nil)
	out := make([]byte, 0, len(nonce)+len(ct))
	out = append(out, nonce...)
	out = append(out, ct...)
	return hex.EncodeToString(out), nil
}

// Decrypt hex(nonce+ciphertext) → 明文。
func (a *AES) Decrypt(cipherHex string) (string, error) {
	raw, err := hex.DecodeString(cipherHex)
	if err != nil {
		return "", err
	}
	if len(raw) < nonceSize+a.gcm.Overhead() {
		return "", errors.New("密文长度异常")
	}
	nonce, ct := raw[:nonceSize], raw[nonceSize:]
	pt, err := a.gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", err
	}
	return string(pt), nil
}