package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	log.Println("正在加载环境变量...")
	if err != nil {
		log.Fatalf("请在程序同级目录创建.env 文件：%s", err)
	}
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/v2", apiHandler)
	log.Println("正在启动服务...")
	info := os.Getenv("HOST") + ":" + os.Getenv("PORT")
	log.Println("服务器将在：" + info + "启动...")
	err = http.ListenAndServe(os.Getenv("HOST")+":"+os.Getenv("PORT"), nil)
	if err != nil {
		log.Println(err.Error())
		return
	}
}
