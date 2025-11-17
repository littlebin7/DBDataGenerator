package storage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
)

var (
	// 默认加密密钥（32字节，AES-256）
	// 生产环境应该从环境变量或配置文件读取
	defaultKey = []byte("DBDataGenerator2024SecretKey32Bytes!!")
)

// getEncryptionKey 获取加密密钥
func getEncryptionKey() []byte {
	// 优先从环境变量读取
	key := os.Getenv("DB_GENERATOR_ENCRYPTION_KEY")
	if key != "" {
		// 如果环境变量是base64编码的，先解码
		if decoded, err := base64.StdEncoding.DecodeString(key); err == nil && len(decoded) == 32 {
			return decoded
		}
		// 否则使用环境变量的值，补齐或截断到32字节
		keyBytes := []byte(key)
		if len(keyBytes) < 32 {
			// 补齐到32字节
			padded := make([]byte, 32)
			copy(padded, keyBytes)
			return padded
		}
		return keyBytes[:32]
	}
	return defaultKey
}

// EncryptPassword 加密密码
func EncryptPassword(password string) (string, error) {
	if password == "" {
		return "", nil
	}

	key := getEncryptionKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("创建加密器失败: %w", err)
	}

	// 使用GCM模式
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("创建GCM失败: %w", err)
	}

	// 生成随机nonce
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("生成nonce失败: %w", err)
	}

	// 加密
	ciphertext := aesGCM.Seal(nonce, nonce, []byte(password), nil)

	// Base64编码
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptPassword 解密密码
func DecryptPassword(encryptedPassword string) (string, error) {
	if encryptedPassword == "" {
		return "", nil
	}

	key := getEncryptionKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("创建解密器失败: %w", err)
	}

	// 使用GCM模式
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("创建GCM失败: %w", err)
	}

	// Base64解码
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedPassword)
	if err != nil {
		// 如果不是base64编码，可能是未加密的旧密码，直接返回
		return encryptedPassword, nil
	}

	// 检查长度
	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		// 可能是未加密的旧密码，直接返回
		return encryptedPassword, nil
	}

	// 提取nonce和密文
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// 解密
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		// 解密失败，可能是未加密的旧密码，直接返回
		return encryptedPassword, nil
	}

	return string(plaintext), nil
}
