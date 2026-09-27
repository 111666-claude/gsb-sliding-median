// Package slidingmedian 计算滑动窗口中位数。
package slidingmedian

// Series 保存数据并统计比较次数。
type Series struct {
	values   []int64
	window   int
	compared int
}

// New 建一个序列。输入会被复制，之后计算不会修改调用方的切片。
func New(values []int64, window int) *Series {
	return &Series{values: append([]int64{}, values...), window: window}
}

// Compared 是比较次数（统计用）。
func (s *Series) Compared() int { return s.compared }

// Medians 返回每个窗口的中位数，顺序与窗口出现顺序一致，窗口数为 n-w+1。
// 奇数窗口取中间值，偶数窗口取两个中间值的平均。
// 窗口 <= 0、窗口大于数据或数据为空时返回空。
//
// 实现用两个堆增量维护：下半大顶堆 lo、上半小顶堆 hi，
// 删除用延迟标记，每次滑动 O(log w)，整体 O(n log w)，内存 O(w)。
func (s *Series) Medians() []float64 {
	n, w := len(s.values), s.window
	if w <= 0 || w > n {
		return nil
	}
	s.compared = 0
	lo := &intHeap{less: func(a, b int64) bool { return a > b }, count: &s.compared} // 大顶堆
	hi := &intHeap{less: func(a, b int64) bool { return a < b }, count: &s.compared} // 小顶堆
	delayedLo := map[int64]int{}
	delayedHi := map[int64]int{}
	loSize, hiSize := 0, 0 // 有效元素个数（不含待删除）

	prune := func(h *intHeap, delayed map[int64]int) {
		for h.len() > 0 {
			top := h.top()
			remaining := delayed[top]
			if remaining == 0 {
				return
			}
			if remaining == 1 {
				delete(delayed, top)
			} else {
				delayed[top] = remaining - 1
			}
			h.pop()
		}
	}
	rebalance := func() {
		switch {
		case loSize > hiSize+1:
			prune(lo, delayedLo)
			v := lo.pop()
			hi.push(v)
			loSize--
			hiSize++
		case loSize < hiSize:
			prune(hi, delayedHi)
			v := hi.pop()
			lo.push(v)
			hiSize--
			loSize++
		}
	}
	add := func(v int64) {
		prune(lo, delayedLo)
		if loSize == 0 || v <= lo.top() {
			lo.push(v)
			loSize++
		} else {
			hi.push(v)
			hiSize++
		}
		rebalance()
	}
	slide := func(out, in int64) {
		prune(lo, delayedLo)
		if out <= lo.top() {
			delayedLo[out]++
			loSize--
		} else {
			delayedHi[out]++
			hiSize--
		}
		prune(lo, delayedLo)
		switch {
		case loSize == 0:
			// lo 为空时用 hi 的堆顶做分类，保持 lo 全 <= hi 的划分。
			prune(hi, delayedHi)
			if hiSize == 0 || in <= hi.top() {
				lo.push(in)
				loSize++
			} else {
				hi.push(in)
				hiSize++
			}
		case in <= lo.top():
			lo.push(in)
			loSize++
		default:
			hi.push(in)
			hiSize++
		}
		rebalance()
	}
	median := func() float64 {
		prune(lo, delayedLo)
		if w%2 == 1 {
			return float64(lo.top())
		}
		prune(hi, delayedHi)
		return (float64(lo.top()) + float64(hi.top())) / 2
	}

	out := make([]float64, 0, n-w+1)
	for i := 0; i < w; i++ {
		add(s.values[i])
	}
	out = append(out, median())
	for start := 1; start+w <= n; start++ {
		slide(s.values[start-1], s.values[start+w-1])
		out = append(out, median())
	}
	return out
}

// intHeap 是按 less 排序的二叉堆，less 为真时 a 更靠近堆顶。
// 每次元素比较都会给 count 加一。
type intHeap struct {
	items []int64
	less  func(a, b int64) bool
	count *int
}

func (h *intHeap) len() int   { return len(h.items) }
func (h *intHeap) top() int64 { return h.items[0] }

func (h *intHeap) push(v int64) {
	h.items = append(h.items, v)
	i := len(h.items) - 1
	for i > 0 {
		parent := (i - 1) / 2
		*h.count++
		if !h.less(h.items[i], h.items[parent]) {
			break
		}
		h.items[i], h.items[parent] = h.items[parent], h.items[i]
		i = parent
	}
}

func (h *intHeap) pop() int64 {
	top := h.items[0]
	n := len(h.items) - 1
	last := h.items[n]
	h.items = h.items[:n]
	if n == 0 {
		return top
	}
	// 沿更靠近堆顶的孩子一路下沉到叶子（每层一次比较），再把 last 上浮到位。
	i := 0
	child := 1
	for child < n {
		if right := child + 1; right < n {
			*h.count++
			if h.less(h.items[right], h.items[child]) {
				child = right
			}
		}
		h.items[i] = h.items[child]
		i = child
		child = 2*i + 1
	}
	for i > 0 {
		parent := (i - 1) / 2
		*h.count++
		if !h.less(last, h.items[parent]) {
			break
		}
		h.items[i] = h.items[parent]
		i = parent
	}
	h.items[i] = last
	return top
}
