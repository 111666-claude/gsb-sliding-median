package slidingmedian

import "sort"

// MediansNaive 是朴素参考实现：每个窗口复制一份再排序，
// 用来和增量实现逐窗口对照。compared 按每个窗口 w 次计（与旧口径一致）。
func MediansNaive(values []int64, window int) (medians []float64, compared int) {
	n := len(values)
	if window <= 0 || window > n {
		return nil, 0
	}
	medians = make([]float64, 0, n-window+1)
	for start := 0; start+window <= n; start++ {
		chunk := append([]int64{}, values[start:start+window]...)
		compared += len(chunk)
		sort.Slice(chunk, func(i, j int) bool { return chunk[i] < chunk[j] })
		if window%2 == 1 {
			medians = append(medians, float64(chunk[window/2]))
		} else {
			medians = append(medians, (float64(chunk[window/2-1])+float64(chunk[window/2]))/2)
		}
	}
	return medians, compared
}
