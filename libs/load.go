package libs

import (
	"HitokotoGo/entity"
	"encoding/json"
	"log"
	"os"
)

func LoadAllSentences(category string) []entity.S {
	const (
		sentencesDir = "data/sentences"
		allCategory  = "all"
	)

	if category == "" || category == allCategory {
		category = allCategory
	}

	var sentences []entity.S

	dirInfo, err := os.Stat(sentencesDir)
	if err != nil || !dirInfo.IsDir() {
		log.Println("句子包目录不存在")
		return sentences
	}

	files, err := os.ReadDir(sentencesDir)
	if err != nil {
		log.Printf("读取句子包目录失败: %v", err)
		return sentences
	}

	targetFileName := category + ".json"
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if category != allCategory && file.Name() != targetFileName {
			continue
		}

		fileSentences, err := loadSentencesFromFile(sentencesDir + "/" + file.Name())
		if err != nil {
			continue
		}
		sentences = append(sentences, fileSentences...)
	}

	return sentences
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
