package main

import (
	"HitokotoGo/entity"
	"HitokotoGo/libs"
	"encoding/json"
	"log"
	"net/http"
	"os"
)

var ALLSentences []entity.S

// statsHandler
// 统计数据展示
func statsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	categories := libs.LoadCategories()
	if categories == nil {
		http.Error(w, "failed to load categories", http.StatusInternalServerError)
		return
	}

	type CategoryStat struct {
		Key   string `json:"key"`
		Name  string `json:"name"`
		Desc  string `json:"desc"`
		Count int    `json:"count"`
	}

	var stats []CategoryStat
	total := 0
	for _, cat := range categories {
		sentences := libs.LoadAllSentences(cat.Key)
		count := len(sentences)
		total += count
		stats = append(stats, CategoryStat{
			Key:   cat.Key,
			Name:  cat.Name,
			Desc:  cat.Desc,
			Count: count,
		})
	}

	version := libs.LoadVersion()
	resp := map[string]interface{}{
		"total":          total,
		"categories":     stats,
		"bundle_version": "",
	}
	if version != nil {
		resp["bundle_version"] = version.BundleVersion
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("error encoding stats response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
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

	if _, err := w.Write(page); err != nil {
		log.Printf("failed to write index response: %v", err)
		return
	}
}

// apiHandler
// 随机获取句子
func resolveCategoryKey(r *http.Request) string {
	categoryKey := r.URL.Query().Get("c")
	if categoryKey == "" {
		return "all"
	}
	return categoryKey
}

// apiHandler
// 随机获取句子
func apiHandler(w http.ResponseWriter, r *http.Request) {
	if sentence := libs.GetRandomSentenceFromCache(); sentence != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err := json.NewEncoder(w).Encode(sentence); err != nil {
			log.Printf("error encoding response: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	if len(ALLSentences) == 0 {
		categoryKey := resolveCategoryKey(r)
		ALLSentences = libs.LoadAllSentences(categoryKey)
	}
	if len(ALLSentences) == 0 {
		http.Error(w, "No sentences available", http.StatusInternalServerError)
		return
	}

	randomSentence := ALLSentences[libs.RandInt(0, len(ALLSentences))]

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(randomSentence); err != nil {
		log.Printf("error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
