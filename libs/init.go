package libs

import (
	"context"
	"log"
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func CheckRedis() bool {
	host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")
	password := os.Getenv("REDIS_PASSWORD")
	dbStr := os.Getenv("REDIS_DB")

	addr := host + ":" + port
	log.Println("正在连接Redis服务: " + addr)

	db, err := strconv.Atoi(dbStr)
	if err != nil {
		log.Fatalf("Redis数据库配置错误: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	defer func() {
		if err := rdb.Close(); err != nil {
			log.Fatalf("Redis连接关闭失败: %v", err)
		}
	}()

	return rdb.Ping(ctx).Err() == nil
}
