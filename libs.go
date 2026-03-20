package main

import (
	"HitokotoGo/entity"
	"encoding/json"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func SentencesLoad(key string) ([]entity.SentencesSimple, error) {
	if key == "" {
		key = "a"
	}
	Files, err := os.DirFS("sentences").Open(key + ".json")
	if err != nil {
		return nil, err
	}
	var sentencesSimple []entity.SentencesSimple
	err = json.NewDecoder(Files).Decode(&sentencesSimple)
	if err != nil {
		return nil, err
	}
	return sentencesSimple, nil
}
func CheckSentencesUpdate() {
	var sentencesVersion entity.SentencesVersion
	SentencesDir := os.DirFS("sentences")
	if _, err := fs.Stat(SentencesDir, "version.json"); err != nil {
		if os.IsNotExist(err) {
			log.Println("未找到句子文件，正在下载...")
			DownloadSentences()
		}
	}
	log.Println("正在检查句子更新...")
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("检查更新失败：环境变量加载失败")
	}
	baseUrl := os.Getenv("BASE_URL")
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	// 获取版本信息
	versionResp, err := client.Get(baseUrl + "/version.json")
	if err != nil {
		log.Fatal("无法下载句子包，请检查网络")
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatal("无法关闭响应体")
		}
	}(versionResp.Body)

	if versionResp.StatusCode != 200 {
		log.Fatal("无法下载句子包，服务器返回错误状态")
	}

	versionContent, err := io.ReadAll(versionResp.Body)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(versionContent, &sentencesVersion)
	if err != nil {
		log.Fatal(err)
	}

	println(sentencesVersion.ProtocolVersion)

}

func DownloadSentences() {

}
