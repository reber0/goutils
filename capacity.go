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
	"strconv"
	"strings"
	"sync"
	"time"
)

// 承载能力探测的默认参数
const (
	CapacityRampStepSec                    = 1.0      // 每档至少持续的秒数（太短会被突发缓冲吸收，假通过）
	CapacityRampMinSamples                 = 40       // 每档最少样本数
	CapacityRampMaxWorkers                 = 128      // 单档最大并发 worker
	CapacityBaselineSamples                = 30       // 低负载基线样本数
	CapacityBaselineConcurrency            = 10       // 基线并发
	CapacityMinRate, CapacityMaxRate       = 20, 2000 // 2000 只是安全阀，实际由拐点定
	CapacityMinThreads, CapacityMaxThreads = 4, 64
	CapacityThreadHeadroom                 = 1.5  // 探测发裸 GET，扫描发带 payload 请求（服务端更慢）
	CapacityAssumedP95ms                   = 100  // 不探测时假设的单请求耗时，用于推导线程数
	CapacitySaturationRatio                = 0.8  // 实测吞吐 / 目标 < 0.8 → 饱和
	CapacityLatencyRatio                   = 2.5  // 本档 p95 > 基线 p95 × 2.5 → 排队拐点
	CapacityTailRatio                      = 8.0  // p99 / p95 > 8 → 长尾
	CapacityErrorRatio                     = 0.02 // 429/5xx 占比 > 2% → 已过头
	CapacitySafeMargin                     = 0.65 // 拐点档 × 0.65 = 安全速率
)

// CapacityRamp 是默认阶梯档位（req/s）。
var CapacityRamp = []int{20, 50, 100, 200, 400, 800, 1600}

var capacityBadStatus = map[int]bool{429: true, 500: true, 502: true, 503: true, 504: true}

// RampStep 是一档的实测曲线。
type RampStep struct {
	Rate    int     `json:"rate"`
	Actual  float64 `json:"actual"`
	P50ms   float64 `json:"p50_ms"`
	P95ms   float64 `json:"p95_ms"`
	P99ms   float64 `json:"p99_ms"`
	Errors  int     `json:"errors"`
	Samples int     `json:"samples"`
	Elapsed float64 `json:"elapsed"`
}

// CapacityResult 是探测结果。
type CapacityResult struct {
	P50ms       float64    `json:"p50_ms"`
	P95ms       float64    `json:"p95_ms"`
	SafeRate    int        `json:"safe_rate"`
	SafeThreads int        `json:"safe_threads"`
	Samples     int        `json:"samples"`
	KneeRate    int        `json:"knee_rate"`
	StopReason  string     `json:"stop_reason"`
	Ramp        []RampStep `json:"ramp"`
}

// ProbeLatency 用默认档位阶梯加压测目标容量，推出安全扫描速率与并发数。
// samples/concurrency 传 0 用默认值（每档最少样本 40 / 单档最大 worker 128）。
// 拿不到任何样本返回 ok=false。
//
// 先 30 个 GET（10 并发）定低负载基线 p95，再按 CapacityRamp 逐档开环定速
// （token bucket 排发射时刻，每档 ≥1s 且 ≥40 样本），记录每档实际完成速率、
// p50/p95/p99、429/5xx 数。任一判据触发即为拐点，取上一档：
//
//	实测吞吐 < 目标 × 0.8   → 服务端饱和（主判据）
//	本档 p95 > 基线 × 2.5  → 队列开始堆积
//	p99/p95 > 8            → 长尾，worker 池被占满
//	429/5xx 占比 > 2%      → 已过头（辅助）
//
// SafeRate = 拐点档 × 0.65，夹在 [20, 2000]，再受线程上限压一层；
// SafeThreads = ceil(SafeRate × p95ms / 1000 × 1.5)，夹在 [4, 64]。
//
// 为什么不靠错误码找容量：配置好的站点被打满时通常不返回 5xx/429，只是排队变慢
// （优雅降级）。延迟漂移与吞吐 plateau 才是通用信号。
// 为什么必须开环定速：闭环保不住拐点（coordinated omission）。
// 直连不走代理：要测的是目标本身，代理会把延迟算进基线。
func ProbeLatency(baseURL string, samples, concurrency int) (CapacityResult, bool) {
	return ProbeLatencyWithRamp(baseURL, samples, concurrency, nil)
}

// ProbeLatencyWithRamp 同 ProbeLatency，但可覆盖阶梯档位（nil = 用 CapacityRamp）。
// 对生产站可传 {20, 50, 100} 之类的小档位，避免压测级流量。
func ProbeLatencyWithRamp(baseURL string, samples, concurrency int, ramp []int) (CapacityResult, bool) {
	if baseURL == "" {
		return CapacityResult{}, false
	}
	if !strings.Contains(baseURL, "://") {
		baseURL = "http://" + baseURL // 允许只给 host，探测需要 scheme
	}
	minSamples := samples
	if minSamples <= 0 {
		minSamples = CapacityRampMinSamples
	}
	maxWorkers := concurrency
	if maxWorkers <= 0 {
		maxWorkers = CapacityRampMaxWorkers
	}
	tiers := ramp
	if len(tiers) == 0 {
		tiers = CapacityRamp
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		// 连接池要能容纳最多 worker：默认 MaxIdleConnsPerHost=2，worker 多了每次新建连接，
		// 好站会被误判成「吞吐饱和」。
		MaxIdleConns:        maxWorkers,
		MaxIdleConnsPerHost: maxWorkers,
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
	probeOnce(client, baseURL)

	// 基线延迟必须实测：档位 worker 数若要靠「假设 100ms」推，遇到 500ms 的慢接口
	// 会给出不够的并发，客户端先撑不住 → 好站被误判饱和。
	baselineP95, ok := baselineP95(client, baseURL, CapacityBaselineSamples, CapacityBaselineConcurrency)
	if !ok {
		return CapacityResult{}, false
	}

	res := CapacityResult{}
	var lastGood *RampStep
	stopReason := "全部档位通过"

	for _, targetRate := range tiers {
		count := int(float64(targetRate) * CapacityRampStepSec)
		if count < minSamples {
			count = minSamples
		}
		// 并发只要够把目标速率发出去：max(基线, 上一档实测) × rate × 1.5。
		// 开环定速下多给 worker 不会加压（多出的只空转等 slot），给少了才会假饱和。
		assumeMs := baselineP95
		if n := len(res.Ramp); n > 0 && res.Ramp[n-1].P95ms > assumeMs {
			assumeMs = res.Ramp[n-1].P95ms
		}
		workers := int(math.Ceil(float64(targetRate) * assumeMs / 1000.0 * CapacityThreadHeadroom))
		if workers < 8 {
			workers = 8
		}
		if workers > maxWorkers {
			workers = maxWorkers
		}

		step, ok := rampStep(client, baseURL, targetRate, count, workers)
		if !ok {
			if len(res.Ramp) == 0 {
				return CapacityResult{}, false // 一档都没跑成就当目标连不上
			}
			stopReason = "本档全部失败"
			break
		}
		res.Ramp = append(res.Ramp, step)

		if reason := kneeReason(step, baselineP95); reason != "" {
			if workers >= maxWorkers && step.Actual < float64(step.Rate)*CapacitySaturationRatio {
				reason += "（worker 已到上限 " + itoa(maxWorkers) + "，可能是探测器自身瓶颈而非目标饱和）"
			}
			stopReason = reason
			break
		}
		s := step
		lastGood = &s
	}
	if lastGood == nil {
		// 第一档就触顶（极慢或极弱的站）：按本身延迟退化成保守值
		lastGood = &res.Ramp[0]
	}

	p95 := lastGood.P95ms
	if p95 < 0.001 {
		p95 = 0.001
	}
	hit := float64(lastGood.Rate) * CapacitySafeMargin
	if lastGood.Rate == res.Ramp[0].Rate && len(res.Ramp) == 1 && stopReason != "全部档位通过" {
		hit = float64(lastGood.Rate) // 第一档就饱和，直接用该档
	}
	if hit > CapacityMaxRate {
		hit = CapacityMaxRate
	}
	// 线程能力约束：跑满 rate 需要 rate × p95 个并发，超线程上限就下调 rate，
	// 不给工具跑不满的空头承诺。
	byThreads := int(float64(CapacityMaxThreads) * 1000 / p95 / CapacityThreadHeadroom)
	rate := int(hit)
	if rate > byThreads {
		rate = byThreads
	}
	if rate < CapacityMinRate {
		rate = CapacityMinRate
	}

	res.P50ms = lastGood.P50ms
	res.P95ms = p95
	res.SafeRate = rate
	res.SafeThreads = threadsFor(rate, p95)
	res.KneeRate = lastGood.Rate
	res.StopReason = stopReason
	for _, s := range res.Ramp {
		res.Samples += s.Samples
	}
	return res, true
}

// rampStep 开环定速发 count 个请求：按 1/rate 间隔排发射时刻，worker 只管等到点再发。
func rampStep(client *http.Client, url string, targetRate, count, workers int) (RampStep, bool) {
	var mu sync.Mutex
	issued, bad := 0, 0
	lat := make([]float64, 0, count)
	doneAt := make([]time.Time, 0, count)

	start := time.Now()
	interval := time.Duration(float64(time.Second) / float64(targetRate))

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				mu.Lock()
				if issued >= count {
					mu.Unlock()
					return
				}
				slot := start.Add(time.Duration(issued) * interval)
				issued++
				mu.Unlock()

				if d := time.Until(slot); d > 0 {
					time.Sleep(d)
				}
				ms, status, ok := probeOnceStatus(client, url)
				if !ok {
					mu.Lock()
					bad++
					mu.Unlock()
					continue
				}
				mu.Lock()
				lat = append(lat, ms)
				doneAt = append(doneAt, time.Now())
				if capacityBadStatus[status] {
					bad++
				}
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if len(lat) == 0 {
		return RampStep{}, false
	}
	// 稳态吞吐 = 完成数 / 首尾完成跨度。用「首个请求发出 → 全部结束」会把尾部请求
	// 的完成耗时算进去，短档位被系统性低估 → 好站假饱和。
	elapsed := doneAt[len(doneAt)-1].Sub(doneAt[0]).Seconds()
	if elapsed <= 0 {
		elapsed = 0.001
	}
	sort.Float64s(lat)
	n := len(lat)
	p95 := lat[minInt(n-1, int(float64(n)*0.95))]
	if p95 < 0.001 {
		p95 = 0.001
	}
	actual := float64(n-1) / elapsed
	if n == 1 {
		actual = 1 / elapsed
	}
	return RampStep{
		Rate:    targetRate,
		Actual:  round1(actual),
		P50ms:   round1(lat[n/2]),
		P95ms:   round1(p95),
		P99ms:   round1(lat[minInt(n-1, int(float64(n)*0.99))]),
		Errors:  bad,
		Samples: n,
		Elapsed: round2(elapsed),
	}, true
}

// kneeReason 判断本档是否已到拐点；到了返回原因，否则空串。
func kneeReason(step RampStep, baselineP95 float64) string {
	n := step.Samples
	if n < 1 {
		n = 1
	}
	if float64(step.Errors)/float64(n) > CapacityErrorRatio {
		return itoa(step.Rate) + " rps 出现 " + itoa(step.Errors) + " 个 429/5xx"
	}
	if step.Actual < float64(step.Rate)*CapacitySaturationRatio {
		return itoa(step.Rate) + " rps 只完成 " + ftoa(step.Actual) + " req/s（吞吐饱和）"
	}
	if step.P95ms > baselineP95*CapacityLatencyRatio {
		return itoa(step.Rate) + " rps p95 " + ftoa(step.P95ms) + "ms 超基线 " +
			ftoa(round1(baselineP95)) + "ms 的 " + ftoa(CapacityLatencyRatio) + " 倍"
	}
	if step.P99ms > step.P95ms*CapacityTailRatio {
		return itoa(step.Rate) + " rps p99 " + ftoa(step.P99ms) + "ms 长尾（p95 的 " +
			ftoa(round1(step.P99ms/maxF(step.P95ms, 0.001))) + " 倍）"
	}
	return ""
}

// baselineP95 低负载基线：给定并发送 samples 个 GET，返回 p95。
func baselineP95(client *http.Client, url string, samples, concurrency int) (float64, bool) {
	var mu sync.Mutex
	lat := make([]float64, 0, samples)
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for i := 0; i < samples; i++ {
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			if ms, ok := probeOnce(client, url); ok {
				mu.Lock()
				lat = append(lat, ms)
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if len(lat) == 0 {
		return 0, false
	}
	sort.Float64s(lat)
	n := len(lat)
	p95 := lat[minInt(n-1, int(float64(n)*0.95))]
	if p95 < 0.001 {
		p95 = 0.001
	}
	return p95, true
}

// probeOnce 发一次 GET 返回耗时（毫秒）；连接层失败返回 ok=false。
func probeOnce(client *http.Client, url string) (float64, bool) {
	ms, _, ok := probeOnceStatus(client, url)
	return ms, ok
}

// probeOnceStatus 同 probeOnce，额外返回状态码（用于统计 429/5xx）。
func probeOnceStatus(client *http.Client, url string) (float64, int, bool) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, 0, false
	}
	req.Header.Set("User-Agent", RandomUserAgent())
	t0 := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, false
	}
	io.Copy(io.Discard, io.LimitReader(resp.Body, 2048))
	resp.Body.Close()
	return float64(time.Since(t0).Microseconds()) / 1000.0, resp.StatusCode, true
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
func round2(v float64) float64 { return math.Round(v*100) / 100 }

func maxF(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func itoa(v int) string { return strconv.Itoa(v) }

func ftoa(v float64) string { return strconv.FormatFloat(v, 'f', 1, 64) }
