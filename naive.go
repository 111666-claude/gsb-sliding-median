package slidingmedian

import "sort"

// NaiveMedians 是朴素参考实现：每个窗口都复制出来重新排序后取中位数。
//
// 仅用于正确性对照和比较次数量级对比，时间 O(n·w·log w)、空间 O(w)
// （每窗口一份拷贝）。compared 非空时统计排序中的值比较次数；它不复制
// 之外修改 values。
func NaiveMedians(values []int64, window int, compared *int) []float64 {
	if window <= 0 || window > len(values) {
		return nil
	}
	out := make([]float64, 0, len(values)-window+1)
	for start := 0; start+window <= len(values); start++ {
		chunk := append([]int64(nil), values[start:start+window]...)
		sort.Slice(chunk, func(i, j int) bool {
			if compared != nil {
				*compared++
			}
			return chunk[i] < chunk[j]
		})
		if window&1 == 1 {
			out = append(out, float64(chunk[window/2]))
			continue
		}
		left := chunk[window/2-1]
		right := chunk[window/2]
		out = append(out, float64(left)/2.0+float64(right)/2.0)
	}
	return out
}
