/*
 * @Author: reber
 * @Mail: reber0ask@qq.com
 * @Date: 2022-06-01 23:13:37
 * @LastEditTime: 2025-06-26 13:22:32
 */
package goutils

import (
	"bufio"
	"os"
)

// FileGetContents 获取文件内容
func FileGetContents(filename string) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// FileEachLineRead 按行读取文件内容
func FileEachLineRead(filename string) ([]string, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY, 0664)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var datas []string
	sc := bufio.NewScanner(file)
	for sc.Scan() {
		datas = append(datas, sc.Text())
	}
	return datas, nil
}

// PathExists 判定 文件/文件夹 是否存在
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
