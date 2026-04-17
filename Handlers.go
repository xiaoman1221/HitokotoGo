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
	html := `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<title>Title</title>
	</head>
	<body>
		<h1>This is my first HTML page</h1>
	</body>
	</html>`
	_, err := w.Write([]byte(html))
	if err != nil {
		log.Fatal(err)
		return
	}

}

// apiHandler
// 随机获取句子
func apiHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("c")
	if key == "" {
		categoriesLocal, err := os.ReadFile("data/categories.json")
		if err != nil {
			log.Println("版本文件读取失败,请检查权限")
			key = "a"
		}
		var categories []entity.C
		var categoriesLocalerr = json.Unmarshal(categoriesLocal, &categories)
		if categoriesLocalerr != nil {
			log.Println("分类文件解析失败,请检查权限")
			key = "a"
		}
		key = categories[libs.RandInt(0, len(categories))].Key
	}

	if len(ALLSentences) == 0 {
		ALLSentences = libs.LoadAllSentences(key)
	}

	if len(ALLSentences) == 0 {
		http.Error(w, "No sentences available", http.StatusInternalServerError)
		return
	}

	randomSentence := ALLSentences[libs.RandInt(0, len(ALLSentences))]

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err := json.NewEncoder(w).Encode(randomSentence)
	if err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
