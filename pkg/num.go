/*
 * @Author: reber
 * @Mail: reber0ask@qq.com
 * @Date: 2023-07-12 10:08:10
 * @LastEditTime: 2023-07-12 10:57:26
 */
package goutils

import "github.com/pkg/errors"

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
