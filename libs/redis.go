package libs

import (
	"encoding/json"
	"log"
	"os"
	"strconv"

	"HitokotoGo/entity"

	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client

const sentenceSetKey = "hitokoto:sentences"

func InitRedis() bool {
	host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")
	password := os.Getenv("REDIS_PASSWORD")
	dbStr := os.Getenv("REDIS_DB")

	addr := host + ":" + port
	log.Println("正在连接Redis服务: " + addr)

	db, err := strconv.Atoi(dbStr)
	if err != nil {
		log.Printf("Redis数据库配置错误: %v", err)
		return false
	}

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	if client.Ping(ctx).Err() != nil {
		return false
	}

	rdb = client
	return true
}

func CacheAllSentences(sentences []entity.S) {
	if rdb == nil {
		return
	}
	pipe := rdb.Pipeline()
	for i := range sentences {
		data, err := json.Marshal(sentences[i])
		if err != nil {
			log.Printf("序列化句子失败: %v", err)
			continue
		}
		pipe.SAdd(ctx, sentenceSetKey, data)
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		log.Printf("缓存句子到Redis失败: %v", err)
		return
	}
	log.Printf("已缓存 %d 条句子到Redis", len(sentences))
}

func GetRandomSentenceFromCache() *entity.S {
	if rdb == nil {
		return nil
	}
	data, err := rdb.SRandMember(ctx, sentenceSetKey).Bytes()
	if err != nil {
		return nil
	}
	var s entity.S
	if err := json.Unmarshal(data, &s); err != nil {
		return nil
	}
	return &s
}
