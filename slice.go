/*
 * @Author: reber
 * @Mail: reber0ask@qq.com
 * @Date: 2022-04-28 10:26:09
 * @LastEditTime: 2023-07-28 16:28:54
 */
package goutils

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// SliceListReverse 反转 [][]string
func SliceListReverse(s [][]string) [][]string {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

// SliceToString []string 转为 string
func SliceToString(slc []string) string {
	return "[" + strings.Join(slc, ", ") + "]"
}

// InSlice 判断 needle 是否在 slice, array, map 中
func InSlice(needle interface{}, haystack interface{}) bool {
	// https://github.com/syyongx/php2go/blob/master/php.go#L1265

	val := reflect.ValueOf(haystack)
	switch val.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < val.Len(); i++ {
			if reflect.DeepEqual(needle, val.Index(i).Interface()) {
				return true
			}
		}
	case reflect.Map:
		for _, k := range val.MapKeys() {
			if reflect.DeepEqual(needle, val.MapIndex(k).Interface()) {
				return true
			}
		}
	default:
		panic("haystack: haystack type muset be slice, array or map")
	}

	return false
}

// SortIntSlice []int 排序
func SortIntSlice(t []int) {
	sort.Slice(t, func(i, j int) bool {
		return t[i] < t[j]
	})
}

// SortStringSlice []string 排序
func SortStringSlice(t []string) {
	sort.Slice(t, func(i, j int) bool {
		return t[i] < t[j]
	})
}

// UniqSlice
//
//	对 []int、[]string 去重
func UniqSlice[T comparable](slc []T) []T {
	result := make([]T, 0)
	tmp := make(map[T]bool)
	for _, v := range slc {
		if !tmp[v] {
			tmp[v] = true
			result = append(result, v)
		}
	}
	return result
}

// UniqSlice2D
//
//	对 [][]int、[][]string 去重
func UniqSlice2D[T comparable](slc [][]T) [][]T {
	result := make([][]T, 0)
	tmp := make(map[string]bool) // 使用 string 替代 []T 作为 key
	for _, v := range slc {
		key := ""
		for _, item := range v {
			key += fmt.Sprint(item) + ","
		}
		key = key[:len(key)-1]
		if !tmp[key] {
			tmp[key] = true
			result = append(result, v)
		}
	}
	return result
}
