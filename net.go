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

// RandomUserAgent 随机返回一个 User-Agent。
// 版本号跟主流浏览器走（2026-11：Chrome/Edge 155、Firefox 154、Safari 27）。
// 老版本 UA 是最容易被 WAF 指纹识别的特征，别再加 IE/Chrome 7x 这类历史条目。
func RandomUserAgent() string {
	userAgent := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/155.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/155.0.0.0 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/155.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/155.0.0.0 Safari/537.36 Edg/155.0.0.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/155.0.0.0 Safari/537.36 Edg/155.0.0.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:154.0) Gecko/20100101 Firefox/154.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:154.0) Gecko/20100101 Firefox/154.0",
		"Mozilla/5.0 (X11; Linux x86_64; rv:154.0) Gecko/20100101 Firefox/154.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/27.0 Safari/605.1.15",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 27_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/27.0 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 27_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) CriOS/155.0.0.0 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (iPad; CPU OS 27_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/27.0 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (Linux; Android 16; Pixel 9) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/155.0.0.0 Mobile Safari/537.36",
		"Mozilla/5.0 (Linux; Android 16; SM-S928B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/155.0.0.0 Mobile Safari/537.36 EdgA/155.0.0.0",
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
