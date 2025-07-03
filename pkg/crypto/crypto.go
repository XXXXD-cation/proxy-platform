// Package crypto 提供加密解密和签名验证功能
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
)

// AESCrypto AES加密解密器
type AESCrypto struct {
	key []byte
}

// HMACSigner HMAC签名器
type HMACSigner struct {
	key []byte
}

// NewAESCrypto 创建AES加密解密器
func NewAESCrypto(key []byte) (*AESCrypto, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, fmt.Errorf("AES密钥长度必须是16、24或32字节")
	}

	return &AESCrypto{key: key}, nil
}

// NewHMACSignerFromKey 从密钥创建HMAC签名器
func NewHMACSignerFromKey(key []byte) *HMACSigner {
	return &HMACSigner{key: key}
}

// NewHMACSignerFromString 从字符串密钥创建HMAC签名器
func NewHMACSignerFromString(key string) *HMACSigner {
	return &HMACSigner{key: []byte(key)}
}

// Encrypt 加密数据
func (a *AESCrypto) Encrypt(plaintext []byte) (string, error) {
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return "", fmt.Errorf("创建AES密码器失败: %v", err)
	}

	// 生成随机IV
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", fmt.Errorf("生成IV失败: %v", err)
	}

	// 使用CBC模式加密
	mode := cipher.NewCBCEncrypter(block, iv) // #nosec G407

	// PKCS7填充
	paddedPlaintext := pkcs7Pad(plaintext, aes.BlockSize)

	ciphertext := make([]byte, len(paddedPlaintext))
	mode.CryptBlocks(ciphertext, paddedPlaintext)

	// 将IV和密文组合并编码为base64
	combined := make([]byte, 0, len(iv)+len(ciphertext))
	combined = append(combined, iv...)
	combined = append(combined, ciphertext...)
	return base64.StdEncoding.EncodeToString(combined), nil
}

// Decrypt 解密数据
func (a *AESCrypto) Decrypt(ciphertext string) ([]byte, error) {
	// base64解码
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, fmt.Errorf("base64解码失败: %v", err)
	}

	if len(data) < aes.BlockSize {
		return nil, fmt.Errorf("密文长度不足")
	}

	block, err := aes.NewCipher(a.key)
	if err != nil {
		return nil, fmt.Errorf("创建AES密码器失败: %v", err)
	}

	// 分离IV和密文
	iv := data[:aes.BlockSize]
	cipherData := data[aes.BlockSize:]

	if len(cipherData)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("密文长度不是块大小的倍数")
	}

	// 使用CBC模式解密
	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(cipherData))
	mode.CryptBlocks(plaintext, cipherData)

	// 去除PKCS7填充
	return pkcs7Unpad(plaintext)
}

// Sign 生成HMAC签名
func (h *HMACSigner) Sign(data []byte) string {
	mac := hmac.New(sha256.New, h.key)
	mac.Write(data)
	return hex.EncodeToString(mac.Sum(nil))
}

// Verify 验证HMAC签名
func (h *HMACSigner) Verify(data []byte, signature string) bool {
	expectedSignature := h.Sign(data)
	return hmac.Equal([]byte(expectedSignature), []byte(signature))
}

// GenerateRandomKey 生成指定长度的随机密钥
func GenerateRandomKey(length int) ([]byte, error) {
	key := make([]byte, length)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("生成随机密钥失败: %v", err)
	}
	return key, nil
}

// GenerateRandomString 生成指定长度的随机字符串
func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("生成随机字符串失败: %v", err)
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

// HashPassword 使用SHA256哈希密码
func HashPassword(password, salt string) string {
	hash := sha256.Sum256([]byte(password + salt))
	return hex.EncodeToString(hash[:])
}

// pkcs7Pad 添加PKCS7填充
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := make([]byte, padding)
	for i := range padtext {
		padtext[i] = byte(padding)
	}
	return append(data, padtext...)
}

// pkcs7Unpad 去除PKCS7填充
func pkcs7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("数据为空")
	}

	padding := int(data[len(data)-1])
	if padding > len(data) || padding == 0 {
		return nil, fmt.Errorf("无效的填充")
	}

	for i := len(data) - padding; i < len(data); i++ {
		if data[i] != byte(padding) {
			return nil, fmt.Errorf("无效的填充")
		}
	}

	return data[:len(data)-padding], nil
}
