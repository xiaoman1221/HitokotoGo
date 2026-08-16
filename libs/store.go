package libs

import (
	"HitokotoGo/entity"
	"fmt"
	"log"
	"sync"
)

// 内存句子数据存储：按分类 + uuid 索引，带读写锁保证并发安全。
var (
	storeMu        sync.RWMutex
	categories     []entity.C
	sentencesByKey map[string][]entity.S
	sentenceIndex  map[string]entity.S
)

// ReloadSentences 从磁盘重新加载全部分类与句子数据到内存，并刷新 Redis 缓存。
// 在启动时与每次句子包更新后调用。
func ReloadSentences() error {
	cats := LoadCategories()
	if cats == nil {
		return fmt.Errorf("无法读取分类数据 categories.json")
	}

	byKey := make(map[string][]entity.S, len(cats)+1)
	index := make(map[string]entity.S)
	var all []entity.S
	for _, cat := range cats {
		list := LoadAllSentences(cat.Key)
		byKey[cat.Key] = list
		all = append(all, list...)
	}
	byKey["all"] = all
	for i := range all {
		index[all[i].Uuid] = all[i]
	}

	storeMu.Lock()
	categories = cats
	sentencesByKey = byKey
	sentenceIndex = index
	storeMu.Unlock()

	if rdb == nil && !InitRedis() {
		log.Println("Redis不可用,仅使用内存缓存")
		return nil
	}
	if err := refreshRedisCache(byKey); err != nil {
		log.Printf("Redis缓存刷新失败,仅使用内存缓存: %v", err)
	}
	return nil
}

// GetCategories 返回当前分类列表的副本。
func GetCategories() []entity.C {
	storeMu.RLock()
	defer storeMu.RUnlock()
	if len(categories) == 0 {
		return nil
	}
	out := make([]entity.C, len(categories))
	copy(out, categories)
	return out
}

// IsValidCategory 判断分类 key 是否存在。
func IsValidCategory(key string) bool {
	for _, c := range GetCategories() {
		if c.Key == key {
			return true
		}
	}
	return false
}

// GetSentences 返回指定分类（"" 或 "all" 表示全部分类）的句子列表。
func GetSentences(category string) []entity.S {
	storeMu.RLock()
	defer storeMu.RUnlock()
	if category == "" {
		category = "all"
	}
	return sentencesByKey[category]
}

// GetSentenceByUUID 根据 uuid 查找句子。
func GetSentenceByUUID(uuid string) (entity.S, bool) {
	storeMu.RLock()
	defer storeMu.RUnlock()
	s, ok := sentenceIndex[uuid]
	return s, ok
}

// TotalSentences 返回全部句子数量。
func TotalSentences() int {
	storeMu.RLock()
	defer storeMu.RUnlock()
	return len(sentencesByKey["all"])
}
