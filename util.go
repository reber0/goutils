/*
 * @Author: reber
 * @Mail: reber0ask@qq.com
 * @Date: 2021-11-10 09:48:35
 * @LastEditTime: 2025-06-03 15:19:20
 */

package goutils

import (
	"math/rand"
	"reflect"
	"strconv"
	"time"

	"github.com/nsf/termbox-go"
	"github.com/pkg/errors"
)

// RandomInt 获取区间中的一个随机整数，返回数字范围 [min, max]
func RandomInt(min, max int) int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Intn(max-min+1) + min
}

// RandomString 获取指定长度的随机字符串(数字+大小写字母)
//
//	temStr := RandomString(12)
//	fmt.Println(temStr) // 8Tb7VQqZ5gL4
func RandomString(length int) string {
	charSet := []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	result := make([]byte, length)

	// 创建本地随机生成器（使用时间戳作为随机源）
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)

	for i := range result {
		result[i] = charSet[r.Intn(len(charSet))]
	}
	return string(result)
}

// Str2Unix 将时间字符串转化为东八区时间戳
//
//	timeStr := "2022-06-20 19:52:04"
//	tt := Str2Unix(timeStr)
//	fmt.Println(tt) // 1655725924
func Str2Unix(timeStr string) int64 {
	local, err := time.LoadLocation("Asia/Shanghai") //设置时区
	if err != nil {
		panic(err)
	}

	tt, err := time.ParseInLocation("2006-01-02 15:04:05", timeStr, local)
	if err != nil {
		panic(err)
	}

	return tt.Unix()
}

// Unix2Str 时间戳转时间字符串
//
//	var t1 int = 1655725924
//	var t2 int64 = 1655725924
//	var t3 string = "1655725924"
//	timeStr1, _ := Unix2String(t1)
//	timeStr2, _ := Unix2String(t2)
//	timeStr3, _ := Unix2String(t3)
//	fmt.Println(timeStr1) // "2022-06-20 19:52:04"
//	fmt.Println(timeStr2) // "2022-06-20 19:52:04"
//	fmt.Println(timeStr3) // "2022-06-20 19:52:04"
func Unix2Str(timestamp interface{}) (string, error) {
	tmp := reflect.ValueOf(timestamp)

	var t int64
	switch tmp.Kind() {
	case reflect.Int, reflect.Int64, reflect.Float64:
		t = tmp.Int()
	case reflect.String:
		parsedInt, err := strconv.ParseInt(tmp.String(), 10, 64) // 转为 int64
		if err != nil {
			return "", errors.Errorf("timestamp must be of type int* or float* or string, not %T", timestamp)
		}
		t = parsedInt
	default:
		return "", errors.Errorf("timestamp must be of type int* or float* or string, not %T", timestamp)
	}

	return time.Unix(t, 0).Format("2006-01-02 15:04:05"), nil
}

// GetTermWidth 获取终端宽度
func GetTermWidth() int {
	if err := termbox.Init(); err != nil {
		panic(err)
	}
	width, _ := termbox.Size()
	termbox.Close()

	return width
}

// GetRatio 获取两个 string 的相似度
// https://github.com/syyongx/php2go/blob/master/php.go#L870
func GetRatio(first string, second string) (percent float64) {
	var similarText func(string, string, int, int) int
	similarText = func(str1, str2 string, len1, len2 int) int {
		var sum, max int
		pos1, pos2 := 0, 0

		// Find the longest segment of the same section in two strings
		for i := 0; i < len1; i++ {
			for j := 0; j < len2; j++ {
				for l := 0; (i+l < len1) && (j+l < len2) && (str1[i+l] == str2[j+l]); l++ {
					if l+1 > max {
						max = l + 1
						pos1 = i
						pos2 = j
					}
				}
			}
		}

		if sum = max; sum > 0 {
			if pos1 > 0 && pos2 > 0 {
				sum += similarText(str1, str2, pos1, pos2)
			}
			if (pos1+max < len1) && (pos2+max < len2) {
				s1 := []byte(str1)
				s2 := []byte(str2)
				sum += similarText(string(s1[pos1+max:]), string(s2[pos2+max:]), len1-pos1-max, len2-pos2-max)
			}
		}

		return sum
	}

	l1, l2 := len(first), len(second)
	if l1+l2 == 0 {
		return 0
	}
	sim := similarText(first, second, l1, l2)
	percent = float64(sim*200) / float64(l1+l2)

	return percent / 100
}
