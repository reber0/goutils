/*
 * @Author: reber
 * @Mail: reber0ask@qq.com
 * @Date: 2022-02-21 16:51:19
 * @LastEditTime: 2025-06-03 15:15:03
 */
package goutils

import (
	"fmt"
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
//
//	url, err := goutils.NewURL("https://example.com/path?key=value")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(url.BaseURL())
func NewURL(targetURL string) (*URL, error) {
	urlObj, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	return &URL{
		u: urlObj,
	}, nil
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
	return p.u.Hostname()
}

// Port 获取 Port
func (p *URL) Port() int {
	port := p.u.Port()
	pn, _ := strconv.Atoi(port)
	return pn
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
	MapQuery, err := url.ParseQuery(p.u.RawQuery)
	if err != nil {
		panic(err)
	}
	return MapQuery
}

// Fragment 获取 Fragment
func (p *URL) Fragment() string {
	return p.u.Fragment
}
