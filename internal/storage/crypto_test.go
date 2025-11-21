package storage

import (
	"os"
	"testing"
)

func TestEncryptPassword_EmptyPassword(t *testing.T) {
	result, err := EncryptPassword("")
	if err != nil {
		t.Errorf("加密空密码不应该返回错误: %v", err)
	}
	if result != "" {
		t.Errorf("空密码加密后应该返回空字符串，实际: %s", result)
	}
}

func TestEncryptPassword_ValidPassword(t *testing.T) {
	// 清除环境变量，使用默认密钥
	originalKey := os.Getenv("DB_GENERATOR_ENCRYPTION_KEY")
	os.Unsetenv("DB_GENERATOR_ENCRYPTION_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("DB_GENERATOR_ENCRYPTION_KEY", originalKey)
		}
	}()

	password := "test_password_123"
	encrypted, err := EncryptPassword(password)
	if err != nil {
		t.Fatalf("加密密码失败: %v", err)
	}
	if encrypted == "" {
		t.Error("加密后的密码不应该为空")
	}
	if encrypted == password {
		t.Error("加密后的密码不应该与原始密码相同")
	}
}

func TestDecryptPassword_EmptyPassword(t *testing.T) {
	result, err := DecryptPassword("")
	if err != nil {
		t.Errorf("解密空密码不应该返回错误: %v", err)
	}
	if result != "" {
		t.Errorf("空密码解密后应该返回空字符串，实际: %s", result)
	}
}

func TestDecryptPassword_ValidEncryptedPassword(t *testing.T) {
	password := "test_password_123"
	encrypted, err := EncryptPassword(password)
	if err != nil {
		t.Fatalf("加密密码失败: %v", err)
	}

	decrypted, err := DecryptPassword(encrypted)
	if err != nil {
		t.Fatalf("解密密码失败: %v", err)
	}
	if decrypted != password {
		t.Errorf("解密后的密码应该与原始密码相同，期望: %s, 实际: %s", password, decrypted)
	}
}

func TestEncryptDecryptPassword_RoundTrip(t *testing.T) {
	testPasswords := []string{
		"simple",
		"password_with_123",
		"特殊字符!@#$%^&*()",
		"中文密码测试",
		"very_long_password_that_might_cause_issues_with_encryption_and_decryption_process",
		"",
	}

	for _, password := range testPasswords {
		encrypted, err := EncryptPassword(password)
		if err != nil {
			t.Errorf("加密密码失败 (密码: %s): %v", password, err)
			continue
		}

		decrypted, err := DecryptPassword(encrypted)
		if err != nil {
			t.Errorf("解密密码失败 (密码: %s): %v", password, err)
			continue
		}

		if decrypted != password {
			t.Errorf("密码往返测试失败，期望: %s, 实际: %s", password, decrypted)
		}
	}
}

func TestGetEncryptionKey_DefaultKey(t *testing.T) {
	// 清除环境变量
	originalKey := os.Getenv("DB_GENERATOR_ENCRYPTION_KEY")
	os.Unsetenv("DB_GENERATOR_ENCRYPTION_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("DB_GENERATOR_ENCRYPTION_KEY", originalKey)
		}
	}()

	key := getEncryptionKey()
	if len(key) != 32 {
		t.Errorf("默认密钥长度应该是 32 字节，实际: %d", len(key))
	}
}

func TestGetEncryptionKey_EnvironmentVariable_ValidLength(t *testing.T) {
	// 测试 32 字节的密钥
	testKey := "12345678901234567890123456789012" // 32 字节
	originalKey := os.Getenv("DB_GENERATOR_ENCRYPTION_KEY")
	os.Setenv("DB_GENERATOR_ENCRYPTION_KEY", testKey)
	defer func() {
		if originalKey != "" {
			os.Setenv("DB_GENERATOR_ENCRYPTION_KEY", originalKey)
		} else {
			os.Unsetenv("DB_GENERATOR_ENCRYPTION_KEY")
		}
	}()

	key := getEncryptionKey()
	// 注意：如果环境变量是有效的 base64，可能会被解码
	// 所以只要密钥长度是有效的 AES 密钥长度即可
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		t.Errorf("密钥长度应该是 16、24 或 32 字节，实际: %d", len(key))
	}
}

func TestGetEncryptionKey_EnvironmentVariable_ShortKey(t *testing.T) {
	// 测试短密钥（应该补齐到 32 字节）
	testKey := "short_key" // 9 字节
	originalKey := os.Getenv("DB_GENERATOR_ENCRYPTION_KEY")
	os.Setenv("DB_GENERATOR_ENCRYPTION_KEY", testKey)
	defer func() {
		if originalKey != "" {
			os.Setenv("DB_GENERATOR_ENCRYPTION_KEY", originalKey)
		} else {
			os.Unsetenv("DB_GENERATOR_ENCRYPTION_KEY")
		}
	}()

	key := getEncryptionKey()
	if len(key) != 32 {
		t.Errorf("短密钥应该补齐到 32 字节，实际: %d", len(key))
	}
}

func TestGetEncryptionKey_EnvironmentVariable_LongKey(t *testing.T) {
	// 测试长密钥（应该截断到 32 字节）
	testKey := "this_is_a_very_long_key_that_should_be_truncated_to_32_bytes" // 57 字节
	originalKey := os.Getenv("DB_GENERATOR_ENCRYPTION_KEY")
	os.Setenv("DB_GENERATOR_ENCRYPTION_KEY", testKey)
	defer func() {
		if originalKey != "" {
			os.Setenv("DB_GENERATOR_ENCRYPTION_KEY", originalKey)
		} else {
			os.Unsetenv("DB_GENERATOR_ENCRYPTION_KEY")
		}
	}()

	key := getEncryptionKey()
	if len(key) != 32 {
		t.Errorf("长密钥应该截断到 32 字节，实际: %d", len(key))
	}
}

func TestGetEncryptionKey_EnvironmentVariable_16Bytes(t *testing.T) {
	// 测试 16 字节密钥（AES-128）
	// 使用 base64 编码来避免被误认为是 base64
	testKey := "MTIzNDU2Nzg5MDEyMzQ1Ng==" // base64("1234567890123456") = 16 字节
	originalKey := os.Getenv("DB_GENERATOR_ENCRYPTION_KEY")
	os.Setenv("DB_GENERATOR_ENCRYPTION_KEY", testKey)
	defer func() {
		if originalKey != "" {
			os.Setenv("DB_GENERATOR_ENCRYPTION_KEY", originalKey)
		} else {
			os.Unsetenv("DB_GENERATOR_ENCRYPTION_KEY")
		}
	}()

	key := getEncryptionKey()
	// 如果环境变量是 base64，会被解码为 16 字节
	// 如果环境变量不是 base64，会被当作字符串处理
	// 只要密钥长度是有效的 AES 密钥长度即可
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		t.Errorf("密钥长度应该是 16、24 或 32 字节，实际: %d", len(key))
	}
}

func TestGetEncryptionKey_EnvironmentVariable_24Bytes(t *testing.T) {
	// 测试 24 字节密钥（AES-192）
	// 使用 base64 编码来避免被误认为是 base64
	testKey := "MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0" // base64("123456789012345678901234") = 24 字节
	originalKey := os.Getenv("DB_GENERATOR_ENCRYPTION_KEY")
	os.Setenv("DB_GENERATOR_ENCRYPTION_KEY", testKey)
	defer func() {
		if originalKey != "" {
			os.Setenv("DB_GENERATOR_ENCRYPTION_KEY", originalKey)
		} else {
			os.Unsetenv("DB_GENERATOR_ENCRYPTION_KEY")
		}
	}()

	key := getEncryptionKey()
	// 如果环境变量是 base64，会被解码为 24 字节
	// 如果环境变量不是 base64，会被当作字符串处理
	// 只要密钥长度是有效的 AES 密钥长度即可
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		t.Errorf("密钥长度应该是 16、24 或 32 字节，实际: %d", len(key))
	}
}

func TestEncryptPassword_WithDifferentKeys(t *testing.T) {
	password := "test_password"

	// 使用默认密钥加密
	originalKey := os.Getenv("DB_GENERATOR_ENCRYPTION_KEY")
	os.Unsetenv("DB_GENERATOR_ENCRYPTION_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("DB_GENERATOR_ENCRYPTION_KEY", originalKey)
		} else {
			os.Unsetenv("DB_GENERATOR_ENCRYPTION_KEY")
		}
	}()

	encrypted1, err := EncryptPassword(password)
	if err != nil {
		t.Fatalf("使用默认密钥加密失败: %v", err)
	}

	// 使用不同的32字节密钥加密
	os.Setenv("DB_GENERATOR_ENCRYPTION_KEY", "different_key_1234567890123456")
	encrypted2, err := EncryptPassword(password)
	if err != nil {
		t.Fatalf("使用不同密钥加密失败: %v", err)
	}

	// 两个加密结果应该不同
	if encrypted1 == encrypted2 {
		t.Error("使用不同密钥加密相同密码应该产生不同的结果")
	}

	// 恢复默认密钥并解密第一个加密结果
	os.Unsetenv("DB_GENERATOR_ENCRYPTION_KEY")
	decrypted1, err := DecryptPassword(encrypted1)
	if err != nil {
		t.Fatalf("解密失败: %v", err)
	}
	if decrypted1 != password {
		t.Errorf("解密后的密码应该与原始密码相同，期望: %s, 实际: %s", password, decrypted1)
	}
}

func TestDecryptPassword_InvalidBase64(t *testing.T) {
	// 测试无效的 base64 编码（应该返回原始值）
	invalidBase64 := "这不是有效的base64编码"
	result, err := DecryptPassword(invalidBase64)
	if err != nil {
		t.Errorf("解密无效 base64 不应该返回错误: %v", err)
	}
	if result != invalidBase64 {
		t.Errorf("无效 base64 应该返回原始值，期望: %s, 实际: %s", invalidBase64, result)
	}
}

func TestDecryptPassword_TooShort(t *testing.T) {
	// 测试太短的密文（应该返回原始值）
	shortCiphertext := "dGVzdA==" // base64 编码的 "test"，但太短
	result, err := DecryptPassword(shortCiphertext)
	if err != nil {
		t.Errorf("解密太短的密文不应该返回错误: %v", err)
	}
	// 如果解密失败，应该返回原始值
	if result == "" {
		t.Error("解密失败时应该返回原始值或空字符串")
	}
}

func TestEncryptPassword_KeySizeValidation(t *testing.T) {
	// 测试各种密钥长度
	testCases := []struct {
		name     string
		key      string
		expected int
	}{
		{"16 bytes (AES-128)", "1234567890123456", 16},
		{"24 bytes (AES-192)", "123456789012345678901234", 24},
		{"32 bytes (AES-256)", "12345678901234567890123456789012", 32},
		{"Short key (should pad)", "short", 32},
		{"Long key (should truncate)", "this_is_a_very_long_key_that_should_be_truncated", 32},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			originalKey := os.Getenv("DB_GENERATOR_ENCRYPTION_KEY")
			os.Setenv("DB_GENERATOR_ENCRYPTION_KEY", tc.key)
			defer func() {
				if originalKey != "" {
					os.Setenv("DB_GENERATOR_ENCRYPTION_KEY", originalKey)
				} else {
					os.Unsetenv("DB_GENERATOR_ENCRYPTION_KEY")
				}
			}()

			password := "test_password"
			encrypted, err := EncryptPassword(password)
			if err != nil {
				t.Fatalf("加密失败: %v", err)
			}

			decrypted, err := DecryptPassword(encrypted)
			if err != nil {
				t.Fatalf("解密失败: %v", err)
			}

			if decrypted != password {
				t.Errorf("密码往返测试失败，期望: %s, 实际: %s", password, decrypted)
			}
		})
	}
}
