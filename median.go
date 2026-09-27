// Package slidingmedian 计算滑动窗口中位数。
package slidingmedian

import "sort"

// Series 保存数据并统计比较次数。
type Series struct {
	values   []int64
	window   int
	compared int
}

// New 建一个序列。
func New(values []int64, window int) *Series {
	return &Series{values: append([]int64{}, values...), window: window}
}

// Medians 返回每个窗口的中位数。缺陷：每个窗口都重新排序，且偶数窗口取下中位。
func (s *Series) Medians() []float64 {
	if s.window <= 0 || s.window > len(s.values) {
		return nil
	}
	out := []float64{}
	for start := 0; start+s.window <= len(s.values); start++ {
		chunk := append([]int64{}, s.values[start:start+s.window]...)
		s.compared += len(chunk)
		sort.Slice(chunk, func(i, j int) bool { return chunk[i] < chunk[j] })
		// 缺陷：偶数窗口直接取中间偏左那个
		out = append(out, float64(chunk[s.window/2]))
	}
	return out
}

// Compared 是比较次数（统计用）。
func (s *Series) Compared() int { return s.compared }
