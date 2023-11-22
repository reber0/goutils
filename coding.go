/*
 * @Author: reber
 * @Mail: reber0ask@qq.com
 * @Date: 2022-04-28 09:42:42
 * @LastEditTime: 2023-11-16 19:49:53
 */
package goutils

import (
	"encoding/base64"
	"net/url"
	"regexp"
)

// Base64Encode base64 编码
func Base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// Base64Decode base64 解码
func Base64Decode(data string) []byte {
	plainText, _ := base64.StdEncoding.DecodeString(data)
	return plainText
}

// URLEncode URL 编码
func URLEncode(data string) string {
	escapeURL := url.QueryEscape(data)
	return escapeURL
}

// URLDecode URL 解码
func URLDecode(data string) string {
	enEscapeURL, _ := url.QueryUnescape(data)
	return enEscapeURL
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
