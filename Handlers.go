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
	categoryKey := resolveCategoryKey(r)

	if len(ALLSentences) == 0 {
		ALLSentences = libs.LoadAllSentences(categoryKey)
	}
	sentences := ALLSentences
	if len(sentences) == 0 {
		http.Error(w, "No sentences available", http.StatusInternalServerError)
		return
	}

	randomSentence := sentences[libs.RandInt(0, len(sentences))]

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(randomSentence); err != nil {
		log.Printf("error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
