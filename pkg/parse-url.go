/*
 * @Author: reber
 * @Mail: reber0ask@qq.com
 * @Date: 2022-02-21 16:51:19
 * @LastEditTime: 2023-07-12 10:57:37
 */
package pkg

import (
	"fmt"
	"net"
	"net/url"
	"path"
	"strconv"
	"strings"
)

// URL 解析 url
type URL struct {
	u *url.URL
}

// NewURL 解析 URL
func NewURL(targetURL string) *URL {
	urlObj, _ := url.Parse(targetURL)

	return &URL{
		u: urlObj,
	}
}

// BaseURL 获取 BaseURL
func (p *URL) BaseURL() string {
	return fmt.Sprintf("%s://%s/", p.u.Scheme, p.u.Host)
}

// Scheme 获取 Scheme
func (p *URL) Scheme() string {
	return p.u.Scheme
}

// Username 获取 Username
func (p *URL) Username() string {
	return p.u.User.Username()
}

// Password 获取 Password
func (p *URL) Password() string {
	Pwd, _ := p.u.User.Password()
	return Pwd
}

// Host 获取 Host
func (p *URL) Host() string {
	Host, _, _ := net.SplitHostPort(p.u.Host)
	return Host
}

// Port 获取 Port
func (p *URL) Port() int {
	_, Port, _ := net.SplitHostPort(p.u.Host)
	port, _ := strconv.Atoi(Port)
	return port
}

// Path 获取 Path
func (p *URL) Path() string {
	return p.u.Path
}

// SuffixName 获取 SuffixName
func (p *URL) SuffixName() string {
	fileType := path.Ext(p.u.Path)
	ext := strings.TrimLeft(fileType, ".")

	return ext
}

// RawQuery 获取 RawQuery
func (p *URL) RawQuery() string {
	return p.u.RawQuery
}

// MapQuery 获取 MapQuery
func (p *URL) MapQuery() url.Values {
	MapQuery, _ := url.ParseQuery(p.u.RawQuery)
	return MapQuery
}

// Fragment 获取 Fragment
func (p *URL) Fragment() string {
	return p.u.Fragment
}
