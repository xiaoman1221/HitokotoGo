package main

import (
	"HitokotoGo/libs"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	log.Println("初始化...")
	log.Println("©2026 HitokotoGo Powered By 小满1221")
	log.Println("https://git.1v.fit/xiaoman1221/HitokotoGo")
	log.Println("句子包来自一言：https://hitokoto.cn/")

	// 加载环境变量
	log.Println("正在加载环境变量...")
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("未检测到.env 文件,正在尝试创建！")
		err := godotenv.Write(map[string]string{
			"HOST":               "0.0.0.0",
			"PORT":               "8080",
			"REDIS_HOST":         "localhost",
			"REDIS_PORT":         "6379",
			"REDIS_PASSWORD":     "",
			"REDIS_DB":           "0",
			"SENTENCES_URL":      "https://sentences-bundle.hitokoto.cn",
			"REFRESH_INTERVAL":   "5000",
			"BACKGROUND_REFRESH": "false",
			"AUTO_UPDATE":        "true",
			"BACKGROUND_API":     "https://t.alcy.cc/pc",
		}, ".env")
		if err != nil {
			log.Fatalf("创建.env文件失败！")
		}
		log.Fatalf("创建.env文件成功！请重新启动本程序")
	}

	log.Println("正在检查Redis服务")
	if libs.InitRedis() {
		log.Println("Redis服务正常")
	} else {
		log.Println("Redis服务异常,已回退至文件缓存")
	}

	if libs.CheckSentences() {
		log.Println("句子检查更新完毕")
	} else if !libs.HasLocalData() {
		log.Fatalf("句子检查更新失败,且本地无可用数据,无法启动")
	} else {
		log.Println("句子检查更新失败,将使用本地已有数据启动")
	}

	log.Println("正在加载句子数据...")
	if err := libs.ReloadSentences(); err != nil {
		log.Fatalf("句子数据加载失败: %v", err)
	}
	log.Printf("共加载 %d 条句子", libs.TotalSentences())

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
		if _, err := w.Write([]byte(injectVars(string(page)))); err != nil {
			log.Printf("failed to write stats response: %v", err)
		}
	})
	http.HandleFunc("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend/openapi.yaml")
	})
	http.HandleFunc("/docs", docsHandler)

	if os.Getenv("AUTO_UPDATE") == "true" {
		go startAutoUpdateLoop(time.Hour)
	}

	log.Println("正在启动服务...")
	info := os.Getenv("HOST") + ":" + os.Getenv("PORT")
	log.Println("服务器将在 " + info + " 启动...")

	err = http.ListenAndServe(os.Getenv("HOST")+":"+os.Getenv("PORT"), nil)
	if err != nil {
		log.Println(err.Error())
		return
	}
}

// backgroundAPI 返回可配置的背景图 API，未配置时使用默认值。
func backgroundAPI() string {
	api := os.Getenv("BACKGROUND_API")
	if api == "" {
		return "https://t.alcy.cc/pc"
	}
	return api
}

// injectVars 向静态页面注入公共模板变量（背景图 API）。
func injectVars(content string) string {
	bgAPI, _ := json.Marshal(backgroundAPI())
	content = strings.ReplaceAll(content, "{{BACKGROUND_API}}", string(bgAPI))
	return content
}

func startAutoUpdateLoop(interval time.Duration) {
	if !libs.AutoUpdate() {
		log.Println("自动更新失败，将在下一个运行周期重试")
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		if !libs.AutoUpdate() {
			log.Println("自动更新失败，将在下一个运行周期重试")
		}
	}
}
