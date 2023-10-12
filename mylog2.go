/*
 * @Author: reber
 * @Mail: reber0ask@qq.com
 * @Date: 2023-10-12 17:55:40
 * @LastEditTime: 2023-10-12 17:55:57
 */
package goutils

import (
	"fmt"

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
	fmt.Fprintf(stdout, "%s%s%s\n", red, msg, reset)
}

func Green(msg string) {
	fmt.Fprintf(stdout, "%s%s%s\n", green, msg, reset)
}

func Yellow(msg string) {
	fmt.Fprintf(stdout, "%s%s%s\n", yellow, msg, reset)
}

func Blue(msg string) {
	fmt.Fprintf(stdout, "%s%s%s\n", blue, msg, reset)
}
