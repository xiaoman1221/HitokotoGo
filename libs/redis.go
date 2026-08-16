package libs

import (
	"HitokotoGo/entity"
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

var rdb *redis.Client

const sentenceSetBaseKey = "hitokoto:sentences"

// sentenceKey 返回分类对应的 Redis key："all"/"" 使用基础 key，其余按分类拆分。
func sentenceKey(category string) string {
	if category == "" || category == "all" {
		return sentenceSetBaseKey
	}
	return sentenceSetBaseKey + ":" + category
}

// InitRedis 初始化 Redis 连接并测试连通性。
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
		_ = client.Close()
		return false
	}

	// 关闭旧连接，避免反复初始化造成连接泄漏
	if rdb != nil {
		_ = rdb.Close()
	}
	rdb = client
	return true
}

// CacheSentencesByKey 按分类把句子写入 Redis（每个分类一个 Set）。
func CacheSentencesByKey(byKey map[string][]entity.S) error {
	if rdb == nil {
		return errors.New("Redis 未初始化")
	}
	for category, list := range byKey {
		if len(list) == 0 {
			continue
		}
		pipe := rdb.Pipeline()
		for i := range list {
			data, err := json.Marshal(list[i])
			if err != nil {
				return err
			}
			pipe.SAdd(ctx, sentenceKey(category), data)
		}
		if _, err := pipe.Exec(ctx); err != nil {
			return err
		}
	}
	log.Printf("已缓存 %d 个分类的句子到Redis", len(byKey))
	return nil
}

// refreshRedisCache 只清空本服务自己的 key 后重新写入，
// 避免 FlushDB 清空整个 Redis 库（可能与其他应用共用）。
func refreshRedisCache(byKey map[string][]entity.S) error {
	keys := make([]string, 0, len(byKey))
	for category := range byKey {
		keys = append(keys, sentenceKey(category))
	}
	if err := rdb.Del(ctx, keys...).Err(); err != nil {
		return err
	}
	return CacheSentencesByKey(byKey)
}

// GetRandomSentenceFromCache 从指定分类的 Redis Set 中随机取一条。
func GetRandomSentenceFromCache(category string) *entity.S {
	if rdb == nil {
		return nil
	}
	data, err := rdb.SRandMember(ctx, sentenceKey(category)).Bytes()
	if err != nil {
		return nil
	}
	var s entity.S
	if err := json.Unmarshal(data, &s); err != nil {
		return nil
	}
	return &s
}
