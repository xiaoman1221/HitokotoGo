package libs

import "math/rand"

// RandInt 返回 [min, max) 区间内的随机整数。
// 当 max <= min 时返回 min，避免 rand.Intn 在非正数上 panic。
func RandInt(min, max int) int {
	if max <= min {
		return min
	}
	return min + rand.Intn(max-min)
}
