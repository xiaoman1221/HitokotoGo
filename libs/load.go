package libs

import (
	"HitokotoGo/entity"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

// loadCategorySentences 加载单个分类的句子文件 data/sentences/<key>.json。
func loadCategorySentences(key string) []entity.S {
	fileSentences, err := loadSentencesFromFile(filepath.Join("data", "sentences", key+".json"))
	if err != nil {
		return nil
	}
	return fileSentences
}

func LoadCategories() []entity.C {
	data, err := os.ReadFile("data/categories.json")
	if err != nil {
		log.Printf("读取 categories.json 失败: %v", err)
		return nil
	}

	var categories []entity.C
	if err := json.Unmarshal(data, &categories); err != nil {
		log.Printf("解析 categories.json 失败: %v", err)
		return nil
	}

	return categories
}

func LoadVersion() *entity.V {
	data, err := os.ReadFile("data/version.json")
	if err != nil {
		log.Printf("读取 version.json 失败: %v", err)
		return nil
	}
	var v entity.V
	if err := json.Unmarshal(data, &v); err != nil {
		log.Printf("解析 version.json 失败: %v", err)
		return nil
	}
	return &v
}

func loadSentencesFromFile(filePath string) ([]entity.S, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("读取文件 %s 失败: %v", filePath, err)
		return nil, err
	}

	var fileSentences []entity.S
	if err := json.Unmarshal(data, &fileSentences); err != nil {
		log.Printf("解析文件 %s 失败: %v", filePath, err)
		return nil, err
	}

	return fileSentences, nil
}
