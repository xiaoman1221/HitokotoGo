package main

import (
	"HitokotoGo/entity"
	"HitokotoGo/libs"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
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

func resolveCategoryKey(r *http.Request) string {
	categoryKey := r.URL.Query().Get("c")
	if categoryKey == "" {
		return "all"
	}
	return categoryKey
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

// apiHandler
// 随机获取句子（支持 ?c= 分类筛选）
func apiHandler(w http.ResponseWriter, r *http.Request) {
	trackRequest()
	defer finishRequest()

	categoryKey := resolveCategoryKey(r)

	if categoryKey != "all" && !libs.IsValidCategory(categoryKey) {
		http.Error(w, "unknown category: "+categoryKey, http.StatusBadRequest)
		return
	}

	if sentence := libs.GetRandomSentenceFromCache(categoryKey); sentence != nil {
		writeJSON(w, sentence)
		return
	}

	sentences := libs.GetSentences(categoryKey)
	if len(sentences) == 0 {
		// 分类合法但无数据（数据文件缺失或为空），对客户端而言是"找不到资源"
		http.Error(w, "no sentences in category: "+categoryKey, http.StatusNotFound)
		return
	}

	writeJSON(w, sentences[libs.RandInt(0, len(sentences))])
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("error encoding response: %v", err)
	}
}
