package main

import (
	"HitokotoGo/libs"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	log.Println("初始化...")

	// 加载环境变量
	err := godotenv.Load(".env")
	log.Println("正在加载环境变量...")
	if err != nil {
		log.Println("未检测到.env 文件,正在尝试创建！")
		err := godotenv.Write(map[string]string{
			"HOST":           "0.0.0.0",
			"PORT":           "n8080",
			"REDIS_HOST":     "127.0.0.1",
			"REDIS_PORT":     "6379",
			"REDIS_PASSWORD": "",
			"REDIS_DB":       "0",
			"SENTENCES_URL":  "https://sentences-bundle.hitokoto.cn",
		}, ".env")
		if err != nil {
			log.Fatalf("创建.env文件失败！")
		}
		log.Fatalf("创建.env文件成功！,请重新启动本程序")
	}
	log.Println("正在检查Redis服务")
	if libs.CheckRedis() {
		log.Println("Redis服务正常")
	} else {
		log.Println("Redis服务异常,已回退至文件缓存")
	}
	if libs.CheckSentences() {
		log.Println("句子检查更新完毕")
	} else {
		log.Fatalf("句子检查更新失败")
	}

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/v2", apiHandler)

	log.Println("正在启动服务...")
	info := os.Getenv("HOST") + ":" + os.Getenv("PORT")
	log.Println("服务器将在 " + info + "启动...")

	err = http.ListenAndServe(os.Getenv("HOST")+":"+os.Getenv("PORT"), nil)
	if err != nil {
		log.Println(err.Error())
		return
	}
}
