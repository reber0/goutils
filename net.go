/*
 * @Author: reber
 * @Mail: reber0ask@qq.com
 * @Date: 2022-06-01 23:13:08
 * @LastEditTime: 2025-06-30 10:10:37
 */
package goutils

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/transform"
)

// IsSiteAlive 判断网站是否存活
func IsSiteAlive(url string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; SiteLiveCheck)")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false // 网络级错误（超时/拒绝等）
	}
	defer resp.Body.Close()
	return true // 任何HTTP响应状态码均视为存活
}

// IsPortOpenSyn 通过 SYN 半开扫描判断端口是否 open（需要 root/管理员权限）
//
//	ip := "1.1.1.1"
//	port := "80"
//	isOpen := IsPortOpenSyn(ip, port)
func IsPortOpenSyn(ip, port string) bool {
	dstPort, err := strconv.Atoi(port)
	if err != nil || dstPort < 1 || dstPort > 65535 {
		return false
	}

	dstIP := net.ParseIP(ip)
	if dstIP == nil || dstIP.To4() == nil {
		return false // 仅支持 IPv4
	}

	// 构造 TCP SYN 包
	tcpHeader := makeTCPHeader(uint16(rand.Intn(64512)+1024), uint16(dstPort), rand.Uint32(), 0, 0x02)
	ipHeader := makeIPHeader(dstIP, tcpHeader)
	synPacket := append(ipHeader, tcpHeader...)

	// 发送 SYN 并等待响应
	return sendSynAndWait(dstIP, synPacket, 3*time.Second)
}

// makeTCPHeader 构造 TCP 头部（20 字节，不含选项和数据）
func makeTCPHeader(srcPort, dstPort uint16, seq, ack uint32, flags uint8) []byte {
	header := make([]byte, 20)
	binary.BigEndian.PutUint16(header[0:2], srcPort)
	binary.BigEndian.PutUint16(header[2:4], dstPort)
	binary.BigEndian.PutUint32(header[4:8], seq)
	binary.BigEndian.PutUint32(header[8:12], ack)
	header[12] = 5 << 4 // data offset (5 * 4 = 20 bytes, no options)
	header[13] = flags
	binary.BigEndian.PutUint16(header[14:16], 65535) // window size
	// checksum 和 urgent pointer 留 0，raw socket 层不校验
	return header
}

// makeIPHeader 构造 IPv4 头部（20 字节）
func makeIPHeader(dstIP net.IP, payload []byte) []byte {
	header := make([]byte, 20)
	header[0] = 0x45 // version=4, IHL=5 (20 bytes)
	header[1] = 0x00 // DSCP/ECN
	totalLen := 20 + len(payload)
	binary.BigEndian.PutUint16(header[2:4], uint16(totalLen))
	binary.BigEndian.PutUint16(header[4:6], 0) // identification
	binary.BigEndian.PutUint16(header[6:8], 0) // flags/fragment offset
	header[8] = 64                              // TTL
	header[9] = 6                               // protocol = TCP
	binary.BigEndian.PutUint16(header[10:12], 0) // checksum (OS 填)
	// src IP 留 0.0.0.0，OS 会填
	copy(header[12:16], dstIP.To4())
	return header
}

// sendSynAndWait 发送 SYN 包并等待响应
func sendSynAndWait(dstIP net.IP, synPacket []byte, timeout time.Duration) bool {
	conn, err := net.DialTimeout("ip4:tcp", dstIP.String(), timeout)
	if err != nil {
		return false
	}
	defer conn.Close()

	// 设置写超时
	if err := conn.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
		return false
	}

	// 发送 SYN
	if _, err := conn.Write(synPacket); err != nil {
		return false
	}

	// 等待响应
	if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return false
	}

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			return false
		}

		// 响应数据至少 40 字节（IP 头 20 + TCP 头 20）
		if n < 40 {
			continue
		}

		// 解析 TCP flags（IP 头后第 13 字节）
		tcpFlags := buf[20+13]
		// SYN+ACK = 0x12
		if tcpFlags&0x12 == 0x12 {
			return true
		}
		// RST = 0x04，端口关闭
		if tcpFlags&0x04 != 0 {
			return false
		}
	}
}

// 获取网站 TLS 证书版本
func GetTLSVersion(url string) (string, error) {
	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.TLS == nil {
		return "Not using TLS", nil
	}

	return tls.VersionName(resp.TLS.Version), nil
}

// IsValidIP 判断是否为合法 IP
func IsValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
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

// RandomUserAgent 随机生成 User-Agent
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
