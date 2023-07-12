/*
 * @Author: reber
 * @Mail: reber0ask@qq.com
 * @Date: 2022-06-01 23:13:37
 * @LastEditTime: 2023-07-12 10:14:38
 */
package goutils

import (
	"bufio"
	"os"
)

// FileGetContents 获取文件内容
func FileGetContents(filename string) string {
	content, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	return string(content)
}

// FileEachLineRead 按行读取文件内容
func FileEachLineRead(filename string) []string {
	file, err := os.OpenFile(filename, os.O_RDONLY, 0664)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var datas []string
	sc := bufio.NewScanner(file)
	for sc.Scan() {
		datas = append(datas, sc.Text())
	}
	return datas
}

// IsFileExist 判定文件是否存在
func IsFileExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		if os.IsNotExist(err) {
			return false
		}
		panic(err)
	}
	return true
}
