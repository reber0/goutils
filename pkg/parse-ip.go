/*
 * @Author: reber
 * @Mail: reber0ask@qq.com
 * @Date: 2022-06-20 16:49:14
 * @LastEditTime: 2023-07-12 10:32:21
 */
package pkg

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"regexp"
	"strconv"
	"strings"
)

// ParseIP 解析 ip 为列表（排除了 0 和 255）
// 	target := "1.1.1.1,2.2.2.2-5,3.3.3.3/30"
// 	ips := ParseIP(target)
// 	// ips: [1.1.1.1
// 	// 2.2.2.2 2.2.2.3 2.2.2.4 2.2.2.5
// 	// 3.3.3.1 3.3.3.2 3.3.3.3]
func ParseIP(target string) []string {
	var ips []string

	// template1 := `^([1-9]|[1-9]\d|1\d{2}|2[0-4]\d|25[0-5])(\.(\d|[1-9]\d|1\d{2}|2[0-4]\d|25[0-5])){3}/([89]|1[0-9]|2[0-9]|3[012])$`
	template1 := `^([1-9]|[1-9]\d|1\d{2}|2[0-4]\d|25[0-5])(\.(\d|[1-9]\d|1\d{2}|2[0-4]\d|25[0-5])){3}/(1[6-9]|2[0-9]|3[012])$`
	template2 := `^([1-9]|[1-9]\d|1\d{2}|2[0-4]\d|25[0-5])(\.(\d|[1-9]\d|1\d{2}|2[0-4]\d|25[0-5])){3}-([1-9]|[1-9]\d|1\d{2}|2[0-4]\d|25[0-5])$`
	template3 := `^([1-9]|[1-9]\d|1\d{2}|2[0-4]\d|25[0-5])(\.(\d|[1-9]\d|1\d{2}|2[0-4]\d|25[0-5])){3}$`

	re1 := regexp.MustCompile(template1)
	re2 := regexp.MustCompile(template2)
	re3 := regexp.MustCompile(template3)

	tmpS := strings.Split(strings.ReplaceAll(target, " ", ""), ",")
	for _, s := range tmpS {
		switch true {
		case re1.MatchString(s):
			ips = append(ips, parse1(s)...)
		case re2.MatchString(s):
			reg := regexp.MustCompile(`(\d{1,3}\.\d{1,3}\.\d{1,3}\.)(\d{1,3})-(\d{1,3})`)
			res := reg.FindStringSubmatch(s)
			prev, x, y := res[1], res[2], res[3]
			start, _ := strconv.Atoi(x)
			end, _ := strconv.Atoi(y)
			if start > end {
				fmt.Println("Invalid IP", s)
			} else {
				ips = append(ips, parse2(prev, start, end)...)
			}
		case re3.MatchString(s):
			ips = append(ips, s)
		default:
			fmt.Println("Invalid IP", s)
		}
	}

	return ips
}

// 解析 CIDR 为 ip 列表
func parse1(s string) []string {
	// https://stackoverflow.com/questions/60540465/how-to-list-all-ips-in-a-network
	var ips []string

	// convert string to IPNet struct
	_, ipv4Net, err := net.ParseCIDR(s)
	if err != nil {
		log.Fatal(err)
	}

	// convert IPNet struct mask and address to uint32
	// network is BigEndian
	mask := binary.BigEndian.Uint32(ipv4Net.Mask)
	start := binary.BigEndian.Uint32(ipv4Net.IP)

	// find the final address
	finish := (start & mask) | (mask ^ 0xffffffff)

	// loop through addresses as uint32
	for i := start; i <= finish; i++ {
		// convert back to net.IP
		ip := make(net.IP, 4)
		binary.BigEndian.PutUint32(ip, i)
		// fmt.Println(ip)
		ipStr := ip.String()
		if strings.HasSuffix(ipStr, ".0") || strings.HasSuffix(ipStr, ".255") {
			continue
		}
		ips = append(ips, ipStr)
	}

	return ips
}

// 解析 1.1.1.1-10 为 ip 列表
func parse2(prev string, start, end int) []string {
	var ips []string

	for i := start; i <= end; i++ {
		ip := fmt.Sprintf("%s%d", prev, i)
		ips = append(ips, ip)
	}

	return ips
}
