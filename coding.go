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
	"strings"
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
// 注意：& 必须先编码，否则编码后的实体会被二次编码
func HTMLEntityEncode(data string) string {
	data = strings.ReplaceAll(data, "&", "&amp;") // & 先编码，防止二次编码
	data = strings.ReplaceAll(data, "<", "&lt;")
	data = strings.ReplaceAll(data, ">", "&gt;")
	data = strings.ReplaceAll(data, "'", "&apos;")
	data = strings.ReplaceAll(data, "\"", "&quot;")
	return data
}

// HTMLEntityDecode html 实体解码
// 注意：&amp; 必须最后解码，否则会导致二次解码（如 &amp;lt; → &lt; → <）
func HTMLEntityDecode(data string) string {
	data = strings.ReplaceAll(data, "&lt;", "<")
	data = strings.ReplaceAll(data, "&gt;", ">")
	data = strings.ReplaceAll(data, "&apos;", "'")
	data = strings.ReplaceAll(data, "&quot;", "\"")
	data = strings.ReplaceAll(data, "&amp;", "&") // &amp; 最后解码，防止二次解码
	return data
}
