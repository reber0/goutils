/*
 * @Author: reber
 * @Mail: reber0ask@qq.com
 * @Date: 2022-02-14 14:37:10
 * @LastEditTime: 2025-06-26 16:34:14
 */
package goutils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"time"
)

// Md5 加密
func Md5(plainText []byte) string {
	m := md5.New()
	m.Write(plainText)
	return hex.EncodeToString(m.Sum(nil))
}

// Sha1 加密
func Sha1(plainText []byte) string {
	m := sha1.New()
	m.Write(plainText)
	return hex.EncodeToString(m.Sum(nil))
}

// Sha256 加密
func Sha256(plainText []byte) string {
	m := sha256.New()
	m.Write(plainText)
	return hex.EncodeToString(m.Sum(nil))
}

// Sha512 加密
func Sha512(plainText []byte) string {
	m := sha512.New()
	m.Write(plainText)
	return hex.EncodeToString(m.Sum(nil))
}

// 生成 HMAC-SHA256
func HMACSHA256(data []byte, key []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

// AES-GCM 加密，返回加密后的数据（包含时间戳）
func AESGCMEncrypt(plaintext []byte, key []byte) ([]byte, error) {
	// 准备时间戳数据块
	timestampBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timestampBytes, uint64(time.Now().UnixMilli()))

	// 创建完整数据块: 时间戳 + 实际数据
	fullData := make([]byte, len(timestampBytes)+len(plaintext))
	copy(fullData[:8], timestampBytes)
	copy(fullData[8:], plaintext)

	// 创建 AES 加密块
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 创建 AES-GCM 实例
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// 生成随机 nonce (推荐 12 字节)
	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// 使用 AES-GCM 加密
	ciphertext := aesgcm.Seal(nil, nonce, fullData, nil)

	// 组合 nonce + 密文
	output := make([]byte, len(nonce)+len(ciphertext))
	copy(output, nonce)
	copy(output[len(nonce):], ciphertext)

	return output, nil
}

// AES-GCM 解密，返回解密后的数据和时间戳
func AESGCMDecrypt(ciphertext []byte, key []byte) ([]byte, int64, error) {
	// 创建加密块
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, 0, err
	}

	// 创建 GCM 实例
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, 0, err
	}

	// 分离 nonce 和实际密文
	nonceSize := aesgcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, 0, errors.New("ciphertext too short")
	}

	nonce := ciphertext[:nonceSize]
	ciphertext = ciphertext[nonceSize:]

	// 执行解密并验证认证标签
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, 0, errors.New("decryption failed: likely authentication failure")
	}

	// 5. 提取时间戳
	timestampBytes := plaintext[:8]
	timestamp := int64(binary.BigEndian.Uint64(timestampBytes))

	plaintext = plaintext[8:]

	return plaintext, timestamp, nil
}

// AESCBCEncrypt AES CBC加密，key 的长度必须为 16, 24 或者 32
func AESCBCEncrypt(plainText, key []byte) ([]byte, error) {
	// 验证密钥长度（必须是16/24/32字节）
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		err := errors.New("invalid key length: must be 16, 24, or 32 bytes")
		return nil, err
	}

	// 创建 AES 加密块
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 对明文进行 PKCS7 填充
	plainText = pkcs7Pad(plainText, block.BlockSize())

	// 生成随机的初始向量 IV（长度必须等于块大小，16字节）
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	// 创建CBC加密器
	cipherText := make([]byte, len(plainText))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherText, plainText)

	return append(iv, cipherText...), nil
}

// AESCBCDecrypt AES CBC 解密，key 的长度必须为 16, 24 或者 32
func AESCBCDecrypt(cipherText, key []byte) ([]byte, error) {
	// 验证密钥长度
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		err := errors.New("invalid key size")
		return nil, err
	}

	// 检查密文长度（至少包含一个 IV 块）
	if len(cipherText) < aes.BlockSize {
		err := errors.New("ciphertext too short")
		return nil, err
	}

	// 分离 IV 和实际密文
	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	// 验证密文长度（必须是块大小的整数倍）
	if len(cipherText)%aes.BlockSize != 0 {
		err := errors.New("ciphertext is not a multiple of the block size")
		return nil, err
	}

	// 创建 AES 块
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 创建CBC解密器
	plainText := make([]byte, len(cipherText))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plainText, cipherText)

	// 去除 PKCS7 填充
	plainText, err = pkcs7UnPad(plainText)
	if err != nil {
		return nil, err
	}

	return plainText, nil
}

// PKCS7填充
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// PKCS7去填充
func pkcs7UnPad(data []byte) ([]byte, error) {
	// 空输入检查
	if len(data) == 0 {
		err := errors.New("empty input")
		return nil, err
	}

	// 提取填充值并验证范围
	padding := int(data[len(data)-1])
	if padding < 1 || padding > len(data) {
		err := errors.New("invalid padding")
		return nil, err
	}

	// 验证填充字节是否正确
	for i := 0; i < padding; i++ {
		if data[len(data)-1-i] != byte(padding) {
			err := errors.New("invalid padding")
			return nil, err
		}
	}
	return data[:len(data)-padding], nil
}
