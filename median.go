// Package slidingmedian 计算滑动窗口中位数。
//
// 中位数定义：窗口内元素排序后，奇数窗口取中间值，偶数窗口取两个中间值
// 的平均值，统一返回 float64。
//
// 实现用两个支持按位置删除的二叉堆做增量维护：低半区是大顶堆 lo，
// 高半区是小顶堆 hi；窗口滑动时删除离开窗口的元素、加入进入窗口的元素，
// 每次更新都是 O(log w)，整体 O(n log w)，额外内存 O(w)。
package slidingmedian

// 槽位当前所属的堆。
const (
	slotUnused byte = 0
	slotLo     byte = 1
	slotHi     byte = 2
)

// Series 保存数据并统计比较次数。
//
// New 不复制 values，Medians 也不会修改 values；每次调用都重新构建堆，
// 因此同一输入重复计算结果相同。
type Series struct {
	values   []int64
	window   int
	compared int
}

// New 建一个序列：不复制 values，且不会修改它。
func New(values []int64, window int) *Series {
	return &Series{values: values, window: window}
}

// indexedHeap 是支持按位置 O(log w) 删除的二叉堆。
//
// 每个槽位对应当前窗口里的一个值（环形复用），槽位在堆中的下标记录在
// pos 里，因此删除离开窗口的元素不需要惰性删除，堆里始终只有当前窗口
// 的元素。
type indexedHeap struct {
	values []int64
	pos    []int
	owner  []byte
	which  byte
	slots  []int
	series *Series
}

func (h *indexedHeap) less(a, b int64) bool {
	h.series.compared++
	if h.which == slotLo {
		return a > b // 大顶堆
	}
	return a < b // 小顶堆
}

func (h *indexedHeap) Len() int { return len(h.slots) }

func (h *indexedHeap) swap(i, j int) {
	h.slots[i], h.slots[j] = h.slots[j], h.slots[i]
	h.pos[h.slots[i]] = i
	h.pos[h.slots[j]] = j
}

func (h *indexedHeap) up(i int) {
	for i > 0 {
		parent := (i - 1) / 2
		if !h.less(h.values[h.slots[i]], h.values[h.slots[parent]]) {
			break
		}
		h.swap(i, parent)
		i = parent
	}
}

func (h *indexedHeap) down(i, n int) {
	for {
		left := 2*i + 1
		if left >= n {
			return
		}
		smaller := left
		if right := left + 1; right < n && h.less(h.values[h.slots[right]], h.values[h.slots[left]]) {
			smaller = right
		}
		if !h.less(h.values[h.slots[smaller]], h.values[h.slots[i]]) {
			return
		}
		h.swap(i, smaller)
		i = smaller
	}
}

// push 把槽位 slot 加入堆。
func (h *indexedHeap) push(slot int) {
	h.owner[slot] = h.which
	h.pos[slot] = len(h.slots)
	h.slots = append(h.slots, slot)
	h.up(len(h.slots) - 1)
}

// remove 删除槽位 slot，槽位可立即复用。
func (h *indexedHeap) remove(slot int) {
	i := h.pos[slot]
	last := len(h.slots) - 1
	if i != last {
		h.swap(i, last)
	}
	h.slots = h.slots[:last]
	h.owner[slot] = slotUnused
	if i < last {
		h.down(i, last)
		h.up(i)
	}
}

// top 返回堆顶槽位。
func (h *indexedHeap) top() int { return h.slots[0] }

// Medians 返回每个窗口的中位数：共 n-w+1 个，顺序与窗口出现顺序一致。
// 窗口 <= 0、窗口大于数据或数据为空时返回 nil。
func (s *Series) Medians() []float64 {
	n := len(s.values)
	w := s.window
	if w <= 0 || w > n {
		return nil
	}

	pos := make([]int, w)
	owner := make([]byte, w)
	lo := indexedHeap{values: s.values, pos: pos, owner: owner, which: slotLo, series: s}
	hi := indexedHeap{values: s.values, pos: pos, owner: owner, which: slotHi, series: s}

	rebalance := func() {
		if lo.Len() > hi.Len()+1 {
			slot := lo.top()
			lo.remove(slot)
			hi.push(slot)
		} else if lo.Len() < hi.Len() {
			slot := hi.top()
			hi.remove(slot)
			lo.push(slot)
		}
	}

	median := func() float64 {
		a := s.values[lo.top()]
		if w&1 == 1 {
			return float64(a)
		}
		b := s.values[hi.top()]
		return float64(a)/2.0 + float64(b)/2.0
	}

	insert := func(slot int, value int64) {
		// 与 lo 堆顶比较一次决定归属：不大于堆顶进 lo，否则进 hi；lo 空时直接进 lo。
		if lo.Len() > 0 {
			s.compared++
			if value <= s.values[lo.top()] {
				lo.push(slot)
			} else {
				hi.push(slot)
			}
		} else {
			lo.push(slot)
		}
	}

	out := make([]float64, 0, n-w+1)

	for slot, value := range s.values[:w] {
		insert(slot, value)
		rebalance()
	}
	out = append(out, median())

	for start := 1; start+w <= n; start++ {
		outgoing := (start - 1) % w
		incoming := (start + w - 1) % w
		value := s.values[start+w-1]

		if owner[outgoing] == slotLo {
			lo.remove(outgoing)
		} else {
			hi.remove(outgoing)
		}
		insert(incoming, value)
		rebalance()

		out = append(out, median())
	}

	return out
}

// Compared 返回值比较次数（统计用）。
func (s *Series) Compared() int { return s.compared }
