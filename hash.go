/*
 * @Author: reber
 * @Mail: reber0ask@qq.com
 * @Date: 2022-02-14 14:37:10
 * @LastEditTime: 2025-06-03 15:13:23
 */
package goutils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
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

// AesEncrypt AES 加密，CBC，key 的长度必须为 16, 24 或者 32
func AesEncrypt(plainText, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	blockSize := block.BlockSize()
	plainText = PKCS7Padding(plainText, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	cipherText := make([]byte, len(plainText))
	blockMode.CryptBlocks(cipherText, plainText)
	return cipherText, nil
}

// AesDecrypt AES 解密，CBC，key 的长度必须为 16, 24 或者 32
func AesDecrypt(cipherText, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	plainText := make([]byte, len(cipherText))
	blockMode.CryptBlocks(plainText, cipherText)
	plainText = PKCS7UnPadding(plainText)
	return plainText, nil
}

// PKCS7Padding AES padding
func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

// PKCS7UnPadding AES unpadding
func PKCS7UnPadding(plainText []byte) []byte {
	length := len(plainText)
	unpadding := int(plainText[length-1])
	return plainText[:(length - unpadding)]
}
