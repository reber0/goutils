/*
 * @Author: reber
 * @Mail: reber0ask@qq.com
 * @Date: 2026-11-06
 */
package goutils

import (
	"crypto/tls"
	"io"
	"math"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

// 承载能力探测的默认参数
const (
	CapacitySamples                        = 30
	CapacityConcurrency                    = 10
	CapacityMinRate, CapacityMaxRate       = 20, 200
	CapacityMinThreads, CapacityMaxThreads = 4, 64
	CapacityThreadHeadroom                 = 1.5 // 探测发裸 GET，扫描发带 payload 请求（服务端更慢）
	CapacityAssumedP95ms                   = 100 // 不探测时假设的单请求耗时，用于推导线程数
)

// CapacityResult 是探测结果。
type CapacityResult struct {
	P50ms       float64 `json:"p50_ms"`
	P95ms       float64 `json:"p95_ms"`
	SafeRate    int     `json:"safe_rate"`
	SafeThreads int     `json:"safe_threads"`
	Samples     int     `json:"samples"`
}

// ProbeLatency 发 samples 个 GET 测单请求耗时，推出安全扫描速率与并发数。
// samples/concurrency 传 0 用默认值（30 / 10）。拿不到任何样本返回 ok=false。
//
// SafeRate = 10 × 1000 / P95ms，夹在 [20, 200]：前提是目标约 20 个 worker、
// 每 worker 只用一半容量（10 req/s）。猜小只是扫得慢，猜大站点会挂。
// SafeThreads = ceil(SafeRate × P95ms / 1000 × 1.5)，夹在 [4, 64]：按 Little 定律，
// 线程数只需够把 rate 跑满，1.5 倍余量抵掉裸 GET 与带 payload 请求的延迟差。
//
// 直连不走代理：要测的是目标本身，代理会把延迟算进基线。4xx/5xx 也算有效样本。
func ProbeLatency(baseURL string, samples, concurrency int) (CapacityResult, bool) {
	if baseURL == "" {
		return CapacityResult{}, false
	}
	if !strings.Contains(baseURL, "://") {
		baseURL = "http://" + baseURL // 允许只给 host，探测需要 scheme
	}
	if samples <= 0 {
		samples = CapacitySamples
	}
	if concurrency <= 0 {
		concurrency = CapacityConcurrency
	}

	tr := &http.Transport{
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		MaxIdleConnsPerHost: concurrency,
		Proxy:               nil, // 显式 nil = 不走环境变量代理
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   10 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// 预热一次（串行）：首次 DNS 解析 + TCP/TLS 建连的耗时不计入样本。
	// 不预热时域名目标的 P95 会被冷启动拖高、safe_rate 被严重低估
	// （实测 dvwa.lab 冷启动算出 20，同一目标用 IP 是 200）。
	probeOnce(client, baseURL)

	lat := make([]float64, 0, samples)
	var mu sync.Mutex
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for i := 0; i < samples; i++ {
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			if ms, ok := probeOnce(client, baseURL); ok {
				mu.Lock()
				lat = append(lat, ms) // 4xx/5xx 也算样本
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if len(lat) == 0 {
		return CapacityResult{}, false
	}
	sort.Float64s(lat)
	n := len(lat)
	p50 := lat[n/2]
	idx := int(float64(n) * 0.95)
	if idx > n-1 {
		idx = n - 1
	}
	p95 := lat[idx]
	if p95 < 0.001 {
		p95 = 0.001
	}
	rate := safeRate(p95)
	return CapacityResult{
		P50ms:       round1(p50),
		P95ms:       round1(p95),
		SafeRate:    rate,
		SafeThreads: threadsFor(rate, p95),
		Samples:     n,
	}, true
}

// probeOnce 发一次 GET 返回耗时（毫秒）；连接层失败返回 ok=false。
func probeOnce(client *http.Client, url string) (float64, bool) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, false
	}
	req.Header.Set("User-Agent", RandomUserAgent())
	t0 := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return 0, false
	}
	io.Copy(io.Discard, io.LimitReader(resp.Body, 2048))
	resp.Body.Close()
	return float64(time.Since(t0).Microseconds()) / 1000.0, true
}

// safeRate 按 p95 推安全速率：10 × 1000 / p95，夹在 [20, 200]。
func safeRate(p95ms float64) int {
	r := int(10 * 1000 / p95ms)
	if r < CapacityMinRate {
		return CapacityMinRate
	}
	if r > CapacityMaxRate {
		return CapacityMaxRate
	}
	return r
}

// ThreadsFor 给 rate 返回线程池大小（Little 定律）：并发 = 速率 × 单请求耗时。
// 这里没有实测延迟，按 CapacityAssumedP95ms 假设（100ms）。
// 探测过的场景直接用 ProbeLatency 返回的 SafeThreads，那里用的是实测 p95。
func ThreadsFor(rate int) int {
	return threadsFor(rate, CapacityAssumedP95ms)
}

func threadsFor(rate int, p95ms float64) int {
	t := int(math.Ceil(float64(rate) * p95ms / 1000.0 * CapacityThreadHeadroom))
	if t < CapacityMinThreads {
		return CapacityMinThreads
	}
	if t > CapacityMaxThreads {
		return CapacityMaxThreads
	}
	return t
}

func round1(v float64) float64 { return math.Round(v*10) / 10 }
