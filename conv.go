/*
 * @Author: reber
 * @Mail: reber0ask@qq.com
 * @Date: 2023-11-16 19:44:41
 * @LastEditTime: 2025-06-03 15:10:17
 */
package goutils

import (
	"encoding/hex"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/bitly/go-simplejson"
	"github.com/pkg/errors"
)

// HexEncode 16 进制字符串转 str
func Str2Hex(data string) string {
	return strings.ToUpper(hex.EncodeToString([]byte(data)))
}

// HexDecode str 转 16 进制字符串
func Hex2Str(data string) string {
	decoded, err := hex.DecodeString(data)
	if err != nil {
		panic(err)
	}
	return string(decoded)
}

// Str2Unicode str 转 unicode
func Str2Unicode(sText string) string {
	textQuoted := strconv.QuoteToASCII(sText)
	textUnquoted := textQuoted[1 : len(textQuoted)-1]
	return textUnquoted
}

// Unicode2Str unicode 转 str
func Unicode2Str(raw string) string {
	str, err := strconv.Unquote(strings.Replace(strconv.Quote(raw), `\\u`, `\u`, -1))
	if err != nil {
		panic(err)
	}
	return str
}

// jsonStr 转为 gjson.Result
func JsonStr2Go(jsonStr string) *simplejson.Json {
	sJson, err := simplejson.NewJson([]byte(jsonStr))
	if err != nil {
		panic(err)
	}
	return sJson
}

// Go 格式转为 JsonStr
func Go2JsonStr(goData interface{}) (string, error) {
	bGoData, err := json.Marshal(goData)
	if err != nil {
		panic(err)
	}
	return string(bGoData), nil
}

// Num2Float64 : accept numeric types, return float64-value
func Num2Float64(num interface{}) (float64, error) {
	switch n := num.(type) {
	case float32:
		return float64(n), nil
	case float64:
		return float64(n), nil
	case int:
		return float64(n), nil
	case int16:
		return float64(n), nil
	case int32:
		return float64(n), nil
	case int64:
		return float64(n), nil
	case int8:
		return float64(n), nil
	case uint:
		return float64(n), nil
	case uint16:
		return float64(n), nil
	case uint32:
		return float64(n), nil
	case uint64:
		return float64(n), nil
	case uint8:
		return float64(n), nil
	default:
		return 0.0, errors.Errorf("Num be of type (u)int* or float*, not %T", n)
	}
}
