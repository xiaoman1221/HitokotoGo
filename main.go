package main

import (
	"HitokotoGo/libs"
	"log"
	"mime"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func init() {
	mime.AddExtensionType(".js", "application/javascript")
	mime.AddExtensionType(".css", "text/css")
}

func main() {
	log.Println("初始化...")

	// 加载环境变量
	err := godotenv.Load(".env")
	log.Println("正在加载环境变量...")
	if err != nil {
		log.Println("未检测到.env 文件,正在尝试创建！")
		err := godotenv.Write(map[string]string{
			"HOST":               "0.0.0.0",
			"PORT":               "8080",
			"REDIS_HOST":         "127.0.0.1",
			"REDIS_PORT":         "6379",
			"REDIS_PASSWORD":     "",
			"REDIS_DB":           "0",
			"SENTENCES_URL":      "https://sentences-bundle.hitokoto.cn",
			"REFRESH_INTERVAL":   "5000",
			"BACKGROUND_REFRESH": "false",
		}, ".env")
		if err != nil {
			log.Fatalf("创建.env文件失败！")
		}
		log.Fatalf("创建.env文件成功！,请重新启动本程序")
	}
	log.Println("正在检查Redis服务")
	if libs.InitRedis() {
		log.Println("Redis服务正常")
	} else {
		log.Println("Redis服务异常,已回退至文件缓存")
	}
	if libs.CheckSentences() {
		log.Println("句子检查更新完毕")
	} else {
		log.Fatalf("句子检查更新失败")
	}

	log.Println("正在加载句子数据...")
	ALLSentences = libs.LoadAllSentences("all")
	log.Printf("共加载 %d 条句子", len(ALLSentences))
	libs.CacheAllSentences(ALLSentences)

	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("frontend/css"))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("frontend/js"))))

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/v2", apiHandler)
	http.HandleFunc("/stats/data", statsHandler)
	http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/stats" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		page, err := os.ReadFile("frontend/stats.html")
		if err != nil {
			log.Printf("failed to read stats page: %v", err)
			http.Error(w, "stats page not found", http.StatusInternalServerError)
			return
		}
		if _, err := w.Write(page); err != nil {
			log.Printf("failed to write stats response: %v", err)
		}
	})

	log.Println("正在启动服务...")
	info := os.Getenv("HOST") + ":" + os.Getenv("PORT")
	log.Println("服务器将在 " + info + "启动...")

	err = http.ListenAndServe(os.Getenv("HOST")+":"+os.Getenv("PORT"), nil)
	if err != nil {
		log.Println(err.Error())
		return
	}
}
