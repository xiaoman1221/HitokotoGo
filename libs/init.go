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
	Addr := os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT")
	log.Println("正在连接Redis服务: " + Addr)
	Password := os.Getenv("REDIS_PASSWORD")
	DB, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		log.Fatalf("Redis数据库配置错误: %v", err)
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     Addr,
		Password: Password,
		DB:       DB,
	})
	defer func(rdb *redis.Client) {
		err := rdb.Close()
		if err != nil {
			log.Fatalf("Redis连接关闭失败: %v", err)
		}
	}(rdb)

	err = rdb.Ping(ctx).Err()
	if err != nil {
		return false
	}
	return true
}
