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
	defaultKey = []byte("DBDataGenerator2024SecretKey32!!")
)

// getEncryptionKey 获取加密密钥
// AES 密钥必须是 16、24 或 32 字节（对应 AES-128、AES-192、AES-256）
func getEncryptionKey() []byte {
	// 优先从环境变量读取
	key := os.Getenv("DB_GENERATOR_ENCRYPTION_KEY")
	if key != "" {
		// 如果环境变量是base64编码的，先解码
		if decoded, err := base64.StdEncoding.DecodeString(key); err == nil {
			decodedLen := len(decoded)
			// 检查解码后的长度是否为有效的 AES 密钥长度
			if decodedLen == 16 || decodedLen == 24 || decodedLen == 32 {
				return decoded
			}
			// 如果长度不是标准长度，调整到 32 字节（AES-256）
			if decodedLen < 32 {
				// 补齐到32字节
				padded := make([]byte, 32)
				copy(padded, decoded)
				return padded
			}
			// 截断到32字节
			return decoded[:32]
		}
		// 否则使用环境变量的原始值，补齐或截断到32字节
		keyBytes := []byte(key)
		keyLen := len(keyBytes)

		// 如果长度已经是有效的 AES 密钥长度，直接返回
		if keyLen == 16 || keyLen == 24 || keyLen == 32 {
			return keyBytes
		}

		// 否则调整到 32 字节（AES-256）
		if keyLen < 32 {
			// 补齐到32字节
			padded := make([]byte, 32)
			copy(padded, keyBytes)
			return padded
		}
		// 截断到32字节
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

	// 验证密钥长度（AES 密钥必须是 16、24 或 32 字节）
	keyLen := len(key)
	if keyLen != 16 && keyLen != 24 && keyLen != 32 {
		return "", fmt.Errorf("无效的加密密钥长度: %d 字节（必须是 16、24 或 32 字节）", keyLen)
	}

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

	// 验证密钥长度（AES 密钥必须是 16、24 或 32 字节）
	keyLen := len(key)
	if keyLen != 16 && keyLen != 24 && keyLen != 32 {
		return "", fmt.Errorf("无效的加密密钥长度: %d 字节（必须是 16、24 或 32 字节）", keyLen)
	}

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

// EncryptData 加密数据（用于加密整个文件内容）
func EncryptData(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}

	key := getEncryptionKey()

	// 验证密钥长度
	keyLen := len(key)
	if keyLen != 16 && keyLen != 24 && keyLen != 32 {
		return nil, fmt.Errorf("无效的加密密钥长度: %d 字节（必须是 16、24 或 32 字节）", keyLen)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("创建加密器失败: %w", err)
	}

	// 使用GCM模式
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("创建GCM失败: %w", err)
	}

	// 生成随机nonce
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("生成nonce失败: %w", err)
	}

	// 加密
	ciphertext := aesGCM.Seal(nonce, nonce, data, nil)

	return ciphertext, nil
}

// DecryptData 解密数据（用于解密整个文件内容）
func DecryptData(encryptedData []byte) ([]byte, error) {
	if len(encryptedData) == 0 {
		return encryptedData, nil
	}

	key := getEncryptionKey()

	// 验证密钥长度
	keyLen := len(key)
	if keyLen != 16 && keyLen != 24 && keyLen != 32 {
		return nil, fmt.Errorf("无效的加密密钥长度: %d 字节（必须是 16、24 或 32 字节）", keyLen)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("创建解密器失败: %w", err)
	}

	// 使用GCM模式
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("创建GCM失败: %w", err)
	}

	// 检查长度
	nonceSize := aesGCM.NonceSize()
	if len(encryptedData) < nonceSize {
		return nil, fmt.Errorf("加密数据太短")
	}

	// 提取nonce和密文
	nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]

	// 解密
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("解密失败: %w", err)
	}

	return plaintext, nil
}
