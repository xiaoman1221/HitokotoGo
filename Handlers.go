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
	reqTimestamps  []time.Time
	reqMu          sync.Mutex
)

// reqTimestampMax 触发时间戳裁剪的阈值，防止无人访问统计页时数组无限增长。
const reqTimestampMax = 100000

func trackRequest() {
	totalQueries.Add(1)
	activeRequests.Add(1)
	reqMu.Lock()
	reqTimestamps = append(reqTimestamps, time.Now())
	if len(reqTimestamps) > reqTimestampMax {
		reqTimestamps = trimOldTimestamps(reqTimestamps, time.Now())
	}
	reqMu.Unlock()
}

func finishRequest() {
	activeRequests.Add(-1)
}

// trimOldTimestamps 去掉超过 15 分钟的时间戳（reqTimestamps 按时间有序追加）。
func trimOldTimestamps(ts []time.Time, now time.Time) []time.Time {
	cutoff := now.Add(-15 * time.Minute)
	i := 0
	for i < len(ts) && !ts[i].After(cutoff) {
		i++
	}
	if i == 0 {
		return ts
	}
	return ts[i:]
}

// loadAverages 返回最近 1/5/15 分钟的每分钟请求数均值，以及当前 RPM。
func loadAverages() (load1, load5, load15 float64, rpm int) {
	reqMu.Lock()
	defer reqMu.Unlock()
	now := time.Now()
	cutoff1 := now.Add(-time.Minute)
	cutoff5 := now.Add(-5 * time.Minute)
	cutoff15 := now.Add(-15 * time.Minute)

	var recent []time.Time
	var c1, c5, c15 int
	for _, t := range reqTimestamps {
		if t.After(cutoff15) {
			recent = append(recent, t)
			c15++
			if t.After(cutoff5) {
				c5++
				if t.After(cutoff1) {
					c1++
				}
			}
		}
	}
	reqTimestamps = recent

	load1 = float64(c1)
	load5 = float64(c5) / 5
	load15 = float64(c15) / 15
	rpm = c1
	return
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
	load1, load5, load15, rpm := loadAverages()

	resp := map[string]interface{}{
		"total":           total,
		"categories":      stats,
		"bundle_version":  bundleVersion,
		"total_queries":   totalQueries.Load(),
		"active_requests": activeRequests.Load(),
		"load_1":          load1,
		"load_5":          load5,
		"load_15":         load15,
		"rpm":             rpm,
		"memory_mb":       float64(memStats.Alloc) / 1024 / 1024,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("error encoding stats response: %v", err)
	}
}

// indexHandler
// 首页展示
func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	page, err := os.ReadFile("frontend/index.html")
	if err != nil {
		log.Printf("failed to read index page: %v", err)
		http.Error(w, "index page not found", http.StatusInternalServerError)
		return
	}

	interval := os.Getenv("REFRESH_INTERVAL")
	if interval == "" {
		interval = "5000"
	}
	intervalSeconds := interval
	if n, err := strconv.Atoi(interval); err == nil && n > 0 {
		intervalSeconds = strconv.Itoa((n + 500) / 1000)
	}

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
	page, err := os.ReadFile("frontend/docs.html")
	if err != nil {
		log.Printf("failed to read docs page: %v", err)
		http.Error(w, "docs page not found", http.StatusInternalServerError)
		return
	}
	if _, err := w.Write(page); err != nil {
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
		http.Error(w, "No sentences available", http.StatusInternalServerError)
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
