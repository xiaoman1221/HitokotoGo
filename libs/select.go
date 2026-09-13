package libs

import (
	"HitokotoGo/entity"
	"sort"
)

// SelectRandom 从候选分类中随机返回一句，长度区间为闭区间 [minLength, maxLength]。
// catKeys 为空表示全部分类；未知分类不产生候选，全部无效时返回 false。
// 优先走 Redis 快路径（单分类且长度区间覆盖该分类全部句子），否则在内存中选取。
func SelectRandom(catKeys []string, minLength, maxLength int) (entity.S, bool) {
	keys, ok := normalizeKeys(catKeys)
	if !ok {
		return entity.S{}, false
	}

	// Redis 快路径：单分类且长度区间无需过滤。
	// 刻意不持有 store 锁，避免 Redis 网络 I/O 阻塞句子包热更新。
	if len(keys) == 1 && rdb != nil && minLength <= 0 && maxLength >= CategoryMaxLen(keys[0]) {
		if s := getRandomSentenceFromCache(keys[0]); s != nil {
			return *s, true
		}
	}

	return selectFromMemory(keys, minLength, maxLength)
}

// normalizeKeys 归一化分类列表：空列表视为全部分类，空串视为 all，
// 去重并剔除未知分类。返回 false 表示归一化后无任何有效分类。
func normalizeKeys(catKeys []string) ([]string, bool) {
	storeMu.RLock()
	defer storeMu.RUnlock()

	if len(catKeys) == 0 {
		return []string{"all"}, true
	}
	seen := make(map[string]struct{}, len(catKeys))
	keys := make([]string, 0, len(catKeys))
	for _, k := range catKeys {
		if k == "" {
			k = "all"
		}
		if _, dup := seen[k]; dup {
			continue
		}
		if _, ok := indexByKey[k]; !ok {
			continue
		}
		seen[k] = struct{}{}
		keys = append(keys, k)
	}
	return keys, len(keys) > 0
}

// CategoryMaxLen 返回分类的最大句子长度，未知分类返回 0。
func CategoryMaxLen(key string) int {
	storeMu.RLock()
	defer storeMu.RUnlock()
	if idx, ok := indexByKey[key]; ok {
		return idx.maxLen
	}
	return 0
}

// selectFromMemory 在长度索引上二分选取：单分类直接区间内随机；
// 多分类按各分类命中数加权随机，保证整体均匀。
func selectFromMemory(keys []string, minLength, maxLength int) (entity.S, bool) {
	storeMu.RLock()
	defer storeMu.RUnlock()

	if len(keys) == 1 {
		if _, ok := indexByKey[keys[0]]; !ok {
			// keys 基于旧快照归一化，期间句子包可能已重载并移除该分类
			return entity.S{}, false
		}
		cat := keys[0]
		lo, hi := lengthRange(cat, minLength, maxLength)
		if hi <= lo {
			return entity.S{}, false
		}
		byLength := indexByKey[cat].byLength
		return sentencesByKey[cat][byLength[RandInt(lo, hi)]], true
	}

	total := 0
	type hitRange struct {
		cat    string
		lo, hi int
	}
	ranges := make([]hitRange, 0, len(keys))
	for _, cat := range keys {
		if _, ok := indexByKey[cat]; !ok {
			continue
		}
		lo, hi := lengthRange(cat, minLength, maxLength)
		if hi > lo {
			ranges = append(ranges, hitRange{cat, lo, hi})
			total += hi - lo
		}
	}
	if total == 0 {
		return entity.S{}, false
	}
	r := RandInt(0, total)
	for _, h := range ranges {
		n := h.hi - h.lo
		if r < n {
			byLength := indexByKey[h.cat].byLength
			return sentencesByKey[h.cat][byLength[h.lo+r]], true
		}
		r -= n
	}
	return entity.S{}, false
}

// lengthRange 在按长度排序的下标切片上二分出闭区间 [minLength, maxLength]
// 对应的下标边界 [lo, hi)。
func lengthRange(cat string, minLength, maxLength int) (lo, hi int) {
	list := sentencesByKey[cat]
	byLength := indexByKey[cat].byLength
	lo = sort.Search(len(byLength), func(i int) bool {
		return list[byLength[i]].Length >= minLength
	})
	hi = sort.Search(len(byLength), func(i int) bool {
		return list[byLength[i]].Length > maxLength
	})
	return
}
