/*
 * @Author: reber
 * @Mail: reber0ask@qq.com
 * @Date: 2022-06-01 23:13:08
 * @LastEditTime: 2023-07-12 10:16:24
 */
package pkg

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/go-resty/resty/v2"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/transform"
)

// IsSiteAlive 判断网站是否存活
func IsSiteAlive(url string) bool {
	client := resty.New()
	client.SetTimeout(10 * time.Second)
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	resp, err := client.R().Get(url)
	if err != nil {
		return false
	}
	if resp != nil {
		statusCode := resp.StatusCode()
		if statusCode == 200 {
			return true
		}
	}
	return false
}

// IsValidIP 判断是否为合法 IP
func IsValidIP(ip string) bool {
	if net.ParseIP(ip) != nil {
		return true
	}
	return false
}

// EncodeToUTF8 根据 resty 的 resp 获取 utf-8 编码的 html
func EncodeToUTF8(resp *resty.Response) string {
	body := resp.Body()

	contentType := resp.Header().Get("Content-Type")
	e, name, _ := charset.DetermineEncoding(body, contentType) // 获取编码
	if name != "utf-8" {
		bodyReader := bytes.NewReader(body)
		utf8Obj := transform.NewReader(bodyReader, e.NewDecoder()) // 转化为 utf8 格式
		body, _ := io.ReadAll(utf8Obj)
		return string(body)
	}

	return string(body)
}

// RandomUserAgent 随机生成 X-Forwarded-For
func RandomUserAgent() string {
	userAgent := []string{
		"Mozilla/5.0 (iPhone; CPU iPhone OS 15_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) CriOS/100.0.4896.77 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 15_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) EdgiOS/100.0.1185.50 Version/15.0 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 10_6_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148 OPT/3.2.9",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 12_3_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.3 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 12_6_1 like Mac OS X) AppleWebKit/612.4.9 (KHTML, like Gecko) Mobile/19D52 QHBrowser/2 QihooBrowser/5.2.4",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 15_3_1 like Mac OS X; zh-cn) AppleWebKit/601.1.46 (KHTML, like Gecko) Mobile/19D52 Quark/5.6.5.1336 Mobile",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 15_3_1 like Mac OS X; zh-CN) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/19D52 UCBrowser/13.8.9.1722 Mobile  AliApp(TUnionSDK/0.1.20.4)",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 6_0 like Mac OS X) AppleWebKit/536.26 (KHTML,  like Gecko) Version/6.0 Mobile/10A403 Safari/8536.25",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:91.0) Gecko/20100101 Firefox/91.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/76.0.3809.12 Safari/537.36 OPR/86.0.4363.23 (Edition B2)",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.3 Safari/605.1.15",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_16_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.69 Safari/537.36 QIHU 360EE",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 12_2_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.23 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/93.0.4577.63 Safari/537.36 Edg/100.0.1185.50",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.45 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:99.0) Gecko/20100101 Firefox/99.0",
		"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/99.0.4844.51 Safari/537.36 OPR/86.0.4363.23",
		"Mozilla/5.0 (Windows NT 10.0; WOW64; Trident/7.0; Touch; rv:11.0) like Gecko",
		"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/93.0.4577.15 Safari/537.36 QIHU 360SE",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.16 Safari/537.36",
		"Mozilla/5.0 (X11; Ubuntu; Linux x86 64; rv:79.0) Gecko/20100101 Firefox/79.0",
		"Mozilla/5.0 (Linux; Ubuntu 16.04) AppleWebKit/537.36 Chromium/57.0.2987.110 Safari/537.36",
	}
	return userAgent[RandomInt(0, len(userAgent)-1)]
}

// RandomXFF 随机生成 X-Forwarded-For
func RandomXFF() string {
	int1 := RandomInt(1, 255)
	int2 := RandomInt(1, 255)
	int3 := RandomInt(1, 255)
	int4 := RandomInt(1, 255)
	xff := fmt.Sprintf("%d.%d.%d.%d", int1, int2, int3, int4)
	return xff
}
