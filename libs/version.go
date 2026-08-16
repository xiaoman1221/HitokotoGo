package libs

import (
	"HitokotoGo/entity"
	"encoding/json"
	"fmt"
	"os"
)

// readRemoteVersion 读取远程句子包版本信息。
func readRemoteVersion() (*entity.V, error) {
	sentencesURL := os.Getenv("SENTENCES_URL")
	if sentencesURL == "" {
		return nil, fmt.Errorf("SENTENCES_URL 未配置")
	}
	body, err := httpGet(sentencesURL + "/version.json")
	if err != nil {
		return nil, err
	}
	var v entity.V
	if err := json.Unmarshal(body, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// readLocalVersion 读取本地句子包版本信息。
func readLocalVersion() (*entity.V, error) {
	body, err := os.ReadFile("data/version.json")
	if err != nil {
		return nil, err
	}
	var v entity.V
	if err := json.Unmarshal(body, &v); err != nil {
		return nil, err
	}
	return &v, nil
}
