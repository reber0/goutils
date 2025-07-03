/*
 * @Author: reber
 * @Mail: reber0ask@qq.com
 * @Date: 2022-04-28 09:42:42
 * @LastEditTime: 2025-07-02 16:12:25
 */
package goutils

import (
	"encoding/base64"
	"net/url"
	"regexp"
)

// Base64Encode base64 编码
func Base64Encode[T string | []byte](data T) string {
	switch v := any(data).(type) {
	case string:
		return base64.StdEncoding.EncodeToString([]byte(v))
	case []byte:
		return base64.StdEncoding.EncodeToString(v)
	default:
		return "" // 理论上不会发生
	}
}

// Base64Decode base64 解码
func Base64Decode[T string | []byte](data T) (string, error) {
	var inputBytes []byte

	switch v := any(data).(type) {
	case string:
		inputBytes = []byte(v)
	case []byte:
		inputBytes = v
	}

	// 直接使用 Decode 方法处理字节
	decodedLen := base64.StdEncoding.DecodedLen(len(inputBytes)) // 计算解码后数据的最大长度
	output := make([]byte, decodedLen)                           // 创建足够容量的缓冲区
	n, err := base64.StdEncoding.Decode(output, inputBytes)
	if err != nil {
		return "", err
	}

	// 返回实际解码长度的字符串
	return string(output[:n]), nil
}

// URLEncode URL 编码
func URLEncode(data string) string {
	escapeURL := url.QueryEscape(data)
	return escapeURL
}

// URLDecode URL 解码
func URLDecode(data string) (string, error) {
	enEscapeURL, err := url.QueryUnescape(data)
	if err != nil {
		return "", err
	}
	return enEscapeURL, nil
}

// HTMLEntityEncode html 实体编码
func HTMLEntityEncode(data string) string {
	reg1 := regexp.MustCompile(`&`)
	reg2 := regexp.MustCompile(`<`)
	reg3 := regexp.MustCompile(`>`)
	reg4 := regexp.MustCompile(`'`)
	reg5 := regexp.MustCompile(`"`)
	data = reg1.ReplaceAllString(data, "&amp;")
	data = reg2.ReplaceAllString(data, "&lt;")
	data = reg3.ReplaceAllString(data, "&gt;")
	data = reg4.ReplaceAllString(data, "&apos;")
	data = reg5.ReplaceAllString(data, "&quot;")
	return data
}

// HTMLEntityDecode html 实体解码
func HTMLEntityDecode(data string) string {
	reg1 := regexp.MustCompile(`&amp;`)
	reg2 := regexp.MustCompile(`&lt;`)
	reg3 := regexp.MustCompile(`&gt;`)
	reg4 := regexp.MustCompile(`&apos;`)
	reg5 := regexp.MustCompile(`&quot;`)
	data = reg1.ReplaceAllString(data, "&")
	data = reg2.ReplaceAllString(data, "<")
	data = reg3.ReplaceAllString(data, ">")
	data = reg4.ReplaceAllString(data, "'")
	data = reg5.ReplaceAllString(data, "\"")
	return data
}
