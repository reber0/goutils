/*
 * @Author: reber
 * @Mail: reber0ask@qq.com
 * @Date: 2022-04-28 10:26:09
 * @LastEditTime: 2023-08-02 11:33:30
 */
package goutils

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/pkg/errors"
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

// IsInCol 判断 elem 是否在 collection(slice, array, map) 中
// https://github.com/syyongx/php2go/blob/master/php.go#L1265
func IsInCol(collection interface{}, elem interface{}) bool {
	val := reflect.ValueOf(collection)
	switch val.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < val.Len(); i++ {
			if reflect.DeepEqual(elem, val.Index(i).Interface()) {
				return true
			}
		}
	case reflect.Map:
		for _, k := range val.MapKeys() {
			if reflect.DeepEqual(elem, k.Interface()) {
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

// Num2Float64 : accept numeric types, return float64-value
// https://github.com/signintech/gopdf/blob/master/gopdf.go#L862
func Num2Float64(size interface{}) (fontSize float64, err error) {
	switch size := size.(type) {
	case float32:
		return float64(size), nil
	case float64:
		return float64(size), nil
	case int:
		return float64(size), nil
	case int16:
		return float64(size), nil
	case int32:
		return float64(size), nil
	case int64:
		return float64(size), nil
	case int8:
		return float64(size), nil
	case uint:
		return float64(size), nil
	case uint16:
		return float64(size), nil
	case uint32:
		return float64(size), nil
	case uint64:
		return float64(size), nil
	case uint8:
		return float64(size), nil
	default:
		return 0.0, errors.Errorf("Num be of type (u)int* or float*, not %T", size)
	}
}