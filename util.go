/*
 * @Author: reber
 * @Mail: reber0ask@qq.com
 * @Date: 2021-11-10 09:48:35
 * @LastEditTime: 2025-07-07 17:05:25
 */

package goutils

import (
	"fmt"
	"math/rand"
	"reflect"
	"runtime"
	"strconv"
	"time"

	"github.com/nsf/termbox-go"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
)

// RandomInt 获取区间中的一个随机整数，返回数字范围 [min, max]
func RandomInt(min, max int) int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Intn(max-min+1) + min
}

// RandomString 获取指定长度的随机字符串(数字+大小写字母)
//
//	temStr := RandomString(12)
//	fmt.Println(temStr) // 7U8#+SgVNX+b
func RandomString(length int) string {
	return string(RandomByte(length))
}

// RandomByte 获取指定长度的随机字符(数字+大小写字母)
//
//	temByte := RandomByte(6)
//	fmt.Println(temByte) // [68 106 84 51 64 83]
func RandomByte(length int) []byte {
	digits := "23456789"
	lowerCase := "abcdefghijkmnpqrstuvwxyz"
	upperCase := "ABCDEFGHJKLMNPQRSTUVWXYZ"
	specialChars := "!@#$%^&*_-+="
	allChars := digits + lowerCase + upperCase + specialChars

	r := rand.New(rand.NewSource(time.Now().UnixNano())) // 使用安全的随机数生成器

	// 如果长度小于 4，自动调整为 4
	if length < 4 {
		length = 4
	}

	result := make([]byte, length)

	// 生成必备字符
	result[0] = digits[r.Intn(len(digits))]
	result[1] = lowerCase[r.Intn(len(lowerCase))]
	result[2] = upperCase[r.Intn(len(upperCase))]
	result[3] = specialChars[r.Intn(len(specialChars))]

	// 填充剩余字符
	for i := 4; i < length; i++ {
		result[i] = allChars[r.Intn(len(allChars))]
	}

	// 打乱字符顺序（Fisher-Yates算法）
	for i := length - 1; i > 0; i-- {
		j := r.Intn(i + 1)
		result[i], result[j] = result[j], result[i]
	}

	return result
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

// 获取设备的 HostID、Cup 信息、系统盘容量
func GetDeviceSig() string {
	hostInfo, _ := host.Info()
	hostID := hostInfo.HostID

	cpuSig := ""
	cpus, _ := cpu.Info()
	if len(cpus) > 0 {
		cpuSig = fmt.Sprintf("%s_%d_%s", cpus[0].ModelName, cpus[0].Cores, cpus[0].VendorID)
	}

	arch := runtime.GOARCH

	diskSig := ""
	if sysDiskSize, err := getSystemDiskCapacity(); err == nil {
		diskSig = fmt.Sprintf("%dGB", sysDiskSize/1024/1024/1024)
	}

	composite := fmt.Sprintf("%s|%s|%s|%s", hostID, cpuSig, arch, diskSig)

	return composite
}

func getSystemDiskCapacity() (uint64, error) {
	var mountPoint string
	switch runtime.GOOS {
	case "windows":
		mountPoint = "C:"
	default:
		mountPoint = "/"
	}

	usage, err := disk.Usage(mountPoint)
	if err != nil {
		return 0, err
	}

	return usage.Total, nil
}
