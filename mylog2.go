/*
 * @Author: reber
 * @Mail: reber0ask@qq.com
 * @Date: 2023-10-12 17:55:40
 * @LastEditTime: 2025-07-17 16:51:00
 */
package goutils

import (
	"fmt"
	"strings"
	"time"

	"github.com/mattn/go-colorable"
)

var stdout = colorable.NewColorableStderr()

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
)

func Red(args ...interface{})    { colorPrint(colorRed, args...) }
func Green(args ...interface{})  { colorPrint(colorGreen, args...) }
func Yellow(args ...interface{}) { colorPrint(colorYellow, args...) }
func Blue(args ...interface{})   { colorPrint(colorBlue, args...) }

func colorPrint(colorCode string, args ...interface{}) {
	t := time.Now().Format("15:04:05")
	msg := joinArgs(" ", args...)
	fmt.Fprintf(stdout, "[%s] %s%s%s\n", t, colorCode, msg, colorReset)
}

func joinArgs(sep string, args ...interface{}) string {
	// 处理多参数连接，连接多个参数为字符串
	var b strings.Builder
	for i, arg := range args {
		if i > 0 {
			b.WriteString(sep) // 添加分隔符
		}
		fmt.Fprint(&b, arg) // 格式化为字符串
	}
	msg := b.String()
	return msg
}
