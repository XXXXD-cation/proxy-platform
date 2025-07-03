package crypto

import (
	"bytes"
	"testing"
)

func TestAESCrypto(t *testing.T) {
	// 测试有效的密钥长度
	validKeys := [][]byte{
		make([]byte, 16), // 128-bit
		make([]byte, 24), // 192-bit
		make([]byte, 32), // 256-bit
	}

	for i, key := range validKeys {
		// 填充密钥
		for j := range key {
			key[j] = byte(i + 1)
		}

		aes, err := NewAESCrypto(key)
		if err != nil {
			t.Fatalf("创建AES加密器失败: %v", err)
		}

		// 测试加密解密
		plaintext := []byte("Hello, World! 这是一个测试消息。")

		encrypted, err := aes.Encrypt(plaintext)
		if err != nil {
			t.Fatalf("加密失败: %v", err)
		}

		decrypted, err := aes.Decrypt(encrypted)
		if err != nil {
			t.Fatalf("解密失败: %v", err)
		}

		if !bytes.Equal(plaintext, decrypted) {
			t.Errorf("解密结果不匹配: 期望 %s, 得到 %s", plaintext, decrypted)
		}
	}
}

func TestAESCryptoInvalidKey(t *testing.T) {
	// 测试无效的密钥长度
	invalidKeys := [][]byte{
		make([]byte, 15),
		make([]byte, 17),
		make([]byte, 31),
		make([]byte, 33),
	}

	for _, key := range invalidKeys {
		_, err := NewAESCrypto(key)
		if err == nil {
			t.Errorf("应该返回错误，但没有返回")
		}
	}
}

func TestHMACSigner(t *testing.T) {
	key := []byte("test-hmac-key")
	signer := NewHMACSignerFromKey(key)

	data := []byte("test data for hmac signing")

	// 生成签名
	signature := signer.Sign(data)
	if signature == "" {
		t.Fatalf("签名不能为空")
	}

	// 验证正确的签名
	if !signer.Verify(data, signature) {
		t.Errorf("签名验证失败")
	}

	// 验证错误的签名
	wrongSignature := "wrong_signature"
	if signer.Verify(data, wrongSignature) {
		t.Errorf("错误的签名应该验证失败")
	}

	// 验证被篡改的数据
	tamperedData := []byte("tampered data")
	if signer.Verify(tamperedData, signature) {
		t.Errorf("被篡改的数据应该验证失败")
	}
}

func TestHMACSignerFromString(t *testing.T) {
	key := "test-string-key"
	signer := NewHMACSignerFromString(key)

	data := []byte("test data")
	signature := signer.Sign(data)

	if !signer.Verify(data, signature) {
		t.Errorf("字符串密钥创建的签名器验证失败")
	}
}

func TestGenerateRandomKey(t *testing.T) {
	lengths := []int{16, 24, 32, 64}

	for _, length := range lengths {
		key, err := GenerateRandomKey(length)
		if err != nil {
			t.Fatalf("生成随机密钥失败: %v", err)
		}

		if len(key) != length {
			t.Errorf("密钥长度不正确: 期望 %d, 得到 %d", length, len(key))
		}

		// 生成另一个密钥，确保不同
		key2, err := GenerateRandomKey(length)
		if err != nil {
			t.Fatalf("生成第二个随机密钥失败: %v", err)
		}

		if bytes.Equal(key, key2) {
			t.Errorf("两次生成的密钥相同，随机性不足")
		}
	}
}

func TestGenerateRandomString(t *testing.T) {
	lengths := []int{8, 16, 32, 64}

	for _, length := range lengths {
		str, err := GenerateRandomString(length)
		if err != nil {
			t.Fatalf("生成随机字符串失败: %v", err)
		}

		if len(str) != length {
			t.Errorf("字符串长度不正确: 期望 %d, 得到 %d", length, len(str))
		}

		// 生成另一个字符串，确保不同
		str2, err := GenerateRandomString(length)
		if err != nil {
			t.Fatalf("生成第二个随机字符串失败: %v", err)
		}

		if str == str2 {
			t.Errorf("两次生成的字符串相同，随机性不足")
		}
	}
}

func TestHashPassword(t *testing.T) {
	password := "mypassword"
	salt := "mysalt"

	hash1 := HashPassword(password, salt)
	hash2 := HashPassword(password, salt)

	// 相同输入应该产生相同哈希
	if hash1 != hash2 {
		t.Errorf("相同输入的哈希结果不同")
	}

	// 不同盐应该产生不同哈希
	hash3 := HashPassword(password, "differentsalt")
	if hash1 == hash3 {
		t.Errorf("不同盐的哈希结果相同")
	}

	// 不同密码应该产生不同哈希
	hash4 := HashPassword("differentpassword", salt)
	if hash1 == hash4 {
		t.Errorf("不同密码的哈希结果相同")
	}

	// 哈希长度应该是64字符（SHA256的十六进制表示）
	if len(hash1) != 64 {
		t.Errorf("哈希长度不正确: 期望 64, 得到 %d", len(hash1))
	}
}

func TestAESEncryptionConsistency(t *testing.T) {
	key, err := GenerateRandomKey(32)
	if err != nil {
		t.Fatalf("生成密钥失败: %v", err)
	}

	aes, err := NewAESCrypto(key)
	if err != nil {
		t.Fatalf("创建AES加密器失败: %v", err)
	}

	// 测试多次加密解密
	original := []byte("一致性测试数据")

	for i := 0; i < 10; i++ {
		encrypted, err := aes.Encrypt(original)
		if err != nil {
			t.Fatalf("第%d次加密失败: %v", i, err)
		}

		decrypted, err := aes.Decrypt(encrypted)
		if err != nil {
			t.Fatalf("第%d次解密失败: %v", i, err)
		}

		if !bytes.Equal(original, decrypted) {
			t.Errorf("第%d次加密解密不一致", i)
		}
	}
}

func TestPKCS7PadUnpad(t *testing.T) {
	testCases := [][]byte{
		[]byte(""),
		[]byte("a"),
		[]byte("ab"),
		[]byte("abc"),
		[]byte("abcd"),
		[]byte("abcdefghijklmnop"),  // 正好16字节
		[]byte("abcdefghijklmnopq"), // 17字节
	}

	for _, data := range testCases {
		padded := pkcs7Pad(data, 16)

		// 检查填充后长度是16的倍数
		if len(padded)%16 != 0 {
			t.Errorf("填充后长度不是16的倍数: %d", len(padded))
		}

		unpadded, err := pkcs7Unpad(padded)
		if err != nil {
			t.Errorf("去除填充失败: %v", err)
		}

		if !bytes.Equal(data, unpadded) {
			t.Errorf("填充去填充不一致: 原始 %v, 结果 %v", data, unpadded)
		}
	}
}
