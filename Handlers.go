package main

import (
	"HitokotoGo/entity"
	"HitokotoGo/libs"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
)

var (
	totalQueries   atomic.Int64
	activeRequests atomic.Int64

	reqMu    sync.Mutex
	reqSlots = make(map[int64]int64) // unix 秒 → 该秒请求数，窗口外过期桶会被清理
)

// reqWindowSecs 负载统计窗口（与 load_15 一致）；过期桶清理后 map 条目数有界。
const reqWindowSecs = 15 * 60

// reqSlotsPruneAt 触发清理的条目数阈值（窗口+60s 缓冲的 2 倍），清理均摊 O(1)。
const reqSlotsPruneAt = (reqWindowSecs + 60) * 2

func trackRequest() {
	totalQueries.Add(1)
	activeRequests.Add(1)
	sec := time.Now().Unix()
	reqMu.Lock()
	reqSlots[sec]++
	if len(reqSlots) > reqSlotsPruneAt {
		pruneReqSlots(sec)
	}
	reqMu.Unlock()
}

func finishRequest() {
	activeRequests.Add(-1)
}

// pruneReqSlots 删除统计窗口之外的过期桶，调用方需持有 reqMu。
func pruneReqSlots(nowSec int64) {
	cutoff := nowSec - reqWindowSecs - 60
	for sec := range reqSlots {
		if sec < cutoff {
			delete(reqSlots, sec)
		}
	}
}

// loadAverages 返回最近 1/5/15 分钟的每分钟请求数均值。
func loadAverages() (load1, load5, load15 float64) {
	nowSec := time.Now().Unix()
	reqMu.Lock()
	var c1, c5, c15 int64
	for sec, n := range reqSlots {
		if sec > nowSec-60 {
			c1 += n
		}
		if sec > nowSec-5*60 {
			c5 += n
		}
		if sec > nowSec-reqWindowSecs {
			c15 += n
		}
	}
	pruneReqSlots(nowSec)
	reqMu.Unlock()
	return float64(c1), float64(c5) / 5, float64(c15) / 15
}

// statsHandler
// 统计数据 + 运行状态
func statsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	categories := libs.GetCategories()
	if len(categories) == 0 {
		http.Error(w, "failed to load categories", http.StatusInternalServerError)
		return
	}

	stats := make([]entity.CategoryStat, 0, len(categories))
	total := 0
	for _, cat := range categories {
		count := len(libs.GetSentences(cat.Key))
		total += count
		stats = append(stats, entity.CategoryStat{
			Key:   cat.Key,
			Name:  cat.Name,
			Desc:  cat.Desc,
			Count: count,
		})
	}

	version := libs.LoadVersion()
	bundleVersion := ""
	if version != nil {
		bundleVersion = version.BundleVersion
	}

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	load1, load5, load15 := loadAverages()

	resp := map[string]interface{}{
		"total":           total,
		"categories":      stats,
		"bundle_version":  bundleVersion,
		"total_queries":   totalQueries.Load(),
		"active_requests": activeRequests.Load(),
		"load_1":          load1,
		"load_5":          load5,
		"load_15":         load15,
		"memory_mb":       float64(memStats.Alloc) / 1024 / 1024,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("error encoding stats response: %v", err)
	}
}

// indexHandler
// 首页展示
func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, must-revalidate")

	page, err := os.ReadFile("frontend/index.html")
	if err != nil {
		log.Printf("failed to read index page: %v", err)
		http.Error(w, "index page not found", http.StatusInternalServerError)
		return
	}

	// 校验刷新间隔，保证注入页面 JS 的永远是合法数字字面量
	interval := os.Getenv("REFRESH_INTERVAL")
	n, err := strconv.Atoi(interval)
	if err != nil || n <= 0 {
		n = 5000
		interval = "5000"
	}
	intervalSeconds := strconv.Itoa((n + 500) / 1000)

	bgRefresh := os.Getenv("BACKGROUND_REFRESH")
	if bgRefresh != "true" {
		bgRefresh = "false"
	}

	sentenceJSON := "null"
	noRefresh := "false"
	if uuid := r.URL.Query().Get("uuid"); uuid != "" {
		noRefresh = "true"
		if s, ok := libs.GetSentenceByUUID(uuid); ok {
			data, _ := json.Marshal(s)
			sentenceJSON = string(data)
		}
	}

	content := injectVars(string(page))
	content = strings.ReplaceAll(content, "{{REFRESH_INTERVAL}}", interval)
	content = strings.ReplaceAll(content, "{{REFRESH_INTERVAL_SECONDS}}", intervalSeconds)
	content = strings.ReplaceAll(content, "{{NO_REFRESH}}", noRefresh)
	content = strings.ReplaceAll(content, "{{SENTENCE_JSON}}", sentenceJSON)
	content = strings.ReplaceAll(content, "{{BACKGROUND_REFRESH}}", bgRefresh)

	if _, err := w.Write([]byte(content)); err != nil {
		log.Printf("failed to write index response: %v", err)
	}
}

// 官方一言 API 的默认长度区间（闭区间）。
const (
	defaultMinLength = 0
	defaultMaxLength = 30
)

// callbackNameRe 合法 JSONP 回调名，防止通过回调名注入脚本。
var callbackNameRe = regexp.MustCompile(`^[A-Za-z_$][A-Za-z0-9_$]*(?:\.[A-Za-z_$][A-Za-z0-9_$]*)*$`)

// apiError 按官方一言 API 的错误格式输出（HTTP status + JSON body）。
func apiError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(struct {
		Status  int        `json:"status"`
		Message string     `json:"message"`
		Data    []struct{} `json:"data"`
		TS      int64      `json:"ts"`
	}{Status: status, Message: message, Data: []struct{}{}, TS: time.Now().UnixMilli()})
}

// parseLengthParam 解析长度参数；非法值按官方行为忽略并回退默认值，负值按 0 处理。
func parseLengthParam(raw string, def int) int {
	n, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	if n < 0 {
		return 0
	}
	return n
}

// encodeResponseBody 按 charset 参数转码响应体，返回转码后的字节与 charset 名。
// charset=gbk 时转 GBK，转码失败（含 GBK 外字符）回退 UTF-8。
func encodeResponseBody(data []byte, charset string) ([]byte, string) {
	if charset == "gbk" {
		if out, err := simplifiedchinese.GBK.NewEncoder().Bytes(data); err == nil {
			return out, "gbk"
		}
	}
	return data, "utf-8"
}

// apiHandler 随机获取句子，参数对标官方一言 API（developer.hitokoto.cn/sentence）：
// c（可重复多选）、min_length、max_length、encode（text/json/js）、
// callback（JSONP）、select（配合 js）、charset（utf-8/gbk）。
func apiHandler(w http.ResponseWriter, r *http.Request) {
	trackRequest()
	defer finishRequest()

	query := r.URL.Query()

	minLength := parseLengthParam(query.Get("min_length"), defaultMinLength)
	maxLength := parseLengthParam(query.Get("max_length"), defaultMaxLength)
	if maxLength < minLength {
		apiError(w, http.StatusBadRequest, "`max_length` 不能小于 `min_length`！")
		return
	}

	sentence, ok := libs.SelectRandom(query["c"], minLength, maxLength)
	if !ok {
		apiError(w, http.StatusBadRequest, "很抱歉，没有分类有句子符合长度区间。")
		return
	}

	w.Header().Set("X-Content-Type-Options", "nosniff")
	charset := query.Get("charset")
	switch query.Get("encode") {
	case "text":
		body, cs := encodeResponseBody([]byte(sentence.Hitokoto), charset)
		w.Header().Set("Content-Type", "text/plain; charset="+cs)
		_, _ = w.Write(body)
	case "js":
		writeAPIClientJS(w, sentence.Hitokoto, query.Get("select"), charset)
	default: // json 及其他值（官方行为：回退 JSON）
		writeAPIJSON(w, sentence, query.Get("callback"), charset)
	}
}

// writeAPIClientJS 输出官方 encode=js 格式：自执行函数把句子写入 select 指定的元素。
func writeAPIClientJS(w http.ResponseWriter, text, selector, charset string) {
	if selector == "" {
		selector = ".hitokoto"
	}
	// 选择器与正文以 JSON 字符串字面量内插，避免拼接注入
	sel, _ := json.Marshal(selector)
	txt, _ := json.Marshal(text)
	script := "(function hitokoto(){var hitokoto=" + string(txt) +
		";var dom=document.querySelector(" + string(sel) +
		");Array.isArray(dom)?dom[0].innerText=hitokoto:dom.innerText=hitokoto;})()"
	body, cs := encodeResponseBody([]byte(script), charset)
	w.Header().Set("Content-Type", "application/javascript; charset="+cs)
	_, _ = w.Write(body)
}

// writeAPIJSON 输出 JSON；callback 合法时输出官方 JSONP 格式：;cb("<json>");。
func writeAPIJSON(w http.ResponseWriter, sentence entity.S, callback, charset string) {
	body, err := json.Marshal(sentence)
	if err != nil {
		log.Printf("error encoding sentence: %v", err)
		apiError(w, http.StatusInternalServerError, "服务器繁忙，请稍后再试。")
		return
	}
	contentType := "application/json"
	if callback != "" && callbackNameRe.MatchString(callback) {
		// 官方 JSONP 把整个 JSON 作为字符串字面量传给回调函数
		wrapped, err := json.Marshal(string(body))
		if err != nil {
			apiError(w, http.StatusInternalServerError, "服务器繁忙，请稍后再试。")
			return
		}
		body = []byte(";" + callback + "(" + string(wrapped) + ");")
		contentType = "application/javascript"
	}
	out, cs := encodeResponseBody(body, charset)
	w.Header().Set("Content-Type", contentType+"; charset="+cs)
	_, _ = w.Write(out)
}

// docsHandler
// API 文档页面
func docsHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/docs" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, must-revalidate")
	page, err := os.ReadFile("frontend/docs.html")
	if err != nil {
		log.Printf("failed to read docs page: %v", err)
		http.Error(w, "docs page not found", http.StatusInternalServerError)
		return
	}
	if _, err := w.Write([]byte(injectVars(string(page)))); err != nil {
		log.Printf("failed to write docs response: %v", err)
	}
}
