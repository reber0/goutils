/*
 * @Author: reber
 * @Mail: reber0ask@qq.com
 * @Date: 2023-10-12 17:55:40
 * @LastEditTime: 2023-11-22 16:13:46
 */
package goutils

import (
	"fmt"
	"time"

	"github.com/mattn/go-colorable"
)

var stdout = colorable.NewColorableStderr()

const (
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	blue   = "\033[34m"
	reset  = "\033[0m"
)

func Red(msg string) {
	t := time.Now().Format("15:04:05")
	fmt.Fprintf(stdout, "[%s] %s%s%s\n", t, red, msg, reset)
}

func Green(msg string) {
	t := time.Now().Format("15:04:05")
	fmt.Fprintf(stdout, "[%s] %s%s%s\n", t, green, msg, reset)
}

func Yellow(msg string) {
	t := time.Now().Format("15:04:05")
	fmt.Fprintf(stdout, "[%s] %s%s%s\n", t, yellow, msg, reset)
}

func Blue(msg string) {
	t := time.Now().Format("15:04:05")
	fmt.Fprintf(stdout, "[%s] %s%s%s\n", t, blue, msg, reset)
}
