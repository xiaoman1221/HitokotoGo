package libs

import (
	"HitokotoGo/entity"
	"fmt"
	"log"
	"sort"
	"sync"
)

// 内存句子数据存储：按分类 + uuid 索引，带读写锁保证并发安全。
var (
	storeMu        sync.RWMutex
	categories     []entity.C
	sentencesByKey map[string][]entity.S
	sentenceIndex  map[string]entity.S
	// indexByKey 每个分类的长度索引（按长度升序的下标切片 + 最大句子长度），
	// 支撑 O(log n) 的长度区间随机查询。
	indexByKey map[string]*categoryIndex
)

type categoryIndex struct {
	byLength []int
	maxLen   int
}

// buildLengthIndex 返回按句子长度升序排序的下标切片。
func buildLengthIndex(list []entity.S) []int {
	byLength := make([]int, len(list))
	for i := range byLength {
		byLength[i] = i
	}
	sort.SliceStable(byLength, func(a, b int) bool {
		return list[byLength[a]].Length < list[byLength[b]].Length
	})
	return byLength
}

func maxSentenceLength(list []entity.S) int {
	maxLen := 0
	for i := range list {
		if list[i].Length > maxLen {
			maxLen = list[i].Length
		}
	}
	return maxLen
}

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
		list := loadCategorySentences(cat.Key)
		byKey[cat.Key] = list
		all = append(all, list...)
	}
	byKey["all"] = all
	for i := range all {
		index[all[i].Uuid] = all[i]
	}

	// 为每个分类构建长度索引（含 all）
	idx := make(map[string]*categoryIndex, len(byKey))
	for key, list := range byKey {
		idx[key] = &categoryIndex{byLength: buildLengthIndex(list), maxLen: maxSentenceLength(list)}
	}

	storeMu.Lock()
	// 记录被移除的旧分类，用于清理 Redis 中对应的历史 key
	var removedCats []string
	for _, old := range categories {
		if _, ok := byKey[old.Key]; !ok {
			removedCats = append(removedCats, old.Key)
		}
	}
	categories = cats
	sentencesByKey = byKey
	sentenceIndex = index
	indexByKey = idx
	storeMu.Unlock()

	if rdb == nil && !InitRedis() {
		log.Println("Redis不可用,仅使用内存缓存")
		return nil
	}
	if err := refreshRedisCache(byKey, removedCats...); err != nil {
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

// GetSentences 返回指定分类（"" 或 "all" 表示全部分类）的句子列表。
// 返回的是内部切片，调用方只读，不得修改。
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
