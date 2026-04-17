package libs

import (
	"HitokotoGo/entity"
	"encoding/json"
	"log"
	"os"
)

func LoadAllSentences(c string) []entity.S {
	if c == "" || c == "all" {
		c = "all"
	}
	var sentences []entity.S
	SentencesDir, err := os.Stat("data/sentences")
	if err != nil || !SentencesDir.IsDir() {
		log.Println("句子包目录不存在")
		return sentences
	}

	files, err := os.ReadDir("data/sentences")
	if err != nil {
		log.Printf("读取句子包目录失败: %v", err)
		return sentences
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if c != "all" && file.Name() != c+".json" {
			continue
		}

		filePath := "data/sentences/" + file.Name()
		data, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("读取文件 %s 失败: %v", filePath, err)
			continue
		}

		var fileSentences []entity.S
		err = json.Unmarshal(data, &fileSentences)
		if err != nil {
			log.Printf("解析文件 %s 失败: %v", filePath, err)
			continue
		}

		sentences = append(sentences, fileSentences...)
	}

	return sentences
}
