package slidingmedian

import (
	"math"
	"math/rand"
	"testing"
)

// generate 与命令行入口一致的固定种子数据。
func generate(count int) []int64 {
	out := make([]int64, 0, count)
	state := uint64(20260927)
	for index := 0; index < count; index++ {
		state = state*6364136223846793005 + 1442695040888963407
		out = append(out, int64(state>>55)%1000)
	}
	return out
}

func TestOddWindow(t *testing.T) {
	series := New([]int64{1, 3, 2, 5, 4}, 3)
	medians := series.Medians()
	want := []float64{2, 3, 4}
	if len(medians) != len(want) {
		t.Fatalf("窗口数不对：%v", medians)
	}
	for i := range want {
		if medians[i] != want[i] {
			t.Fatalf("中位数应为 %v，得到 %v", want, medians)
		}
	}
}

func TestEvenWindow(t *testing.T) {
	medians := New([]int64{1, 2, 3, 4}, 4).Medians()
	if len(medians) != 1 || medians[0] != 2.5 {
		t.Fatalf("偶数窗口应取两个中间值的平均 2.5：%v", medians)
	}

	// 逐窗口滑动的偶数窗口。
	got := New([]int64{4, 1, 3, 2, 5}, 2).Medians()
	want := []float64{2.5, 2, 2.5, 3.5}
	if len(got) != len(want) {
		t.Fatalf("窗口数应为 %d：%v", len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("偶数窗口中位数应为 %v，得到 %v", want, got)
		}
	}
}

func TestWindowOne(t *testing.T) {
	got := New([]int64{7, -3, 0}, 1).Medians()
	want := []float64{7, -3, 0}
	if len(got) != len(want) {
		t.Fatalf("窗口为 1 时每个元素都是中位数：%v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("窗口为 1 时中位数应为 %v，得到 %v", want, got)
		}
	}
}

func TestDuplicates(t *testing.T) {
	got := New([]int64{2, 2, 2, 1, 2}, 3).Medians()
	want := []float64{2, 2, 2}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("重复值中位数应为 %v，得到 %v", want, got)
		}
	}
}

func TestInt64Extremes(t *testing.T) {
	// 两个中间值相加会溢出 int64，必须分别转 float64。
	got := New([]int64{math.MaxInt64, math.MinInt64}, 2).Medians()
	if len(got) != 1 {
		t.Fatalf("应有一个窗口：%v", got)
	}
	want := float64(math.MaxInt64)/2.0 + float64(math.MinInt64)/2.0
	if got[0] != want || math.IsNaN(got[0]) {
		t.Fatalf("极值平均应安全计算为 %v，得到 %v", want, got[0])
	}
}

func TestWindowLargerThanData(t *testing.T) {
	if got := New([]int64{1, 2}, 5).Medians(); got != nil {
		t.Fatalf("窗口大于数据应返回空：%v", got)
	}
}

func TestZeroWindow(t *testing.T) {
	if got := New([]int64{1, 2}, 0).Medians(); got != nil {
		t.Fatalf("窗口为 0 应返回空：%v", got)
	}
}

func TestNegativeWindow(t *testing.T) {
	if got := New([]int64{1, 2}, -3).Medians(); got != nil {
		t.Fatalf("窗口为负应返回空：%v", got)
	}
}

func TestEmptyData(t *testing.T) {
	if got := New(nil, 3).Medians(); got != nil {
		t.Fatalf("数据为空应返回空：%v", got)
	}
}

func TestWindowEqualsData(t *testing.T) {
	got := New([]int64{1, 2, 3}, 3).Medians()
	if len(got) != 1 || got[0] != 2 {
		t.Fatalf("窗口等于数据时应只有一个值 2：%v", got)
	}
}

func TestDoesNotModifyInput(t *testing.T) {
	data := generate(500)
	original := append([]int64(nil), data...)
	New(data, 31).Medians()
	for i := range original {
		if data[i] != original[i] {
			t.Fatalf("原地滑动修改了输入数据：位置 %d 从 %d 变为 %d", i, original[i], data[i])
		}
	}
}

func TestRepeatable(t *testing.T) {
	data := generate(1000)
	first := New(data, 17).Medians()
	second := New(data, 17).Medians()
	if len(first) != len(second) {
		t.Fatalf("重复计算长度不一致：%d vs %d", len(first), len(second))
	}
	for i := range first {
		if first[i] != second[i] {
			t.Fatalf("同一输入重复计算结果不一致：位置 %d %v vs %v", i, first[i], second[i])
		}
	}
}

func TestComparedIsCounted(t *testing.T) {
	series := New([]int64{1, 2, 3, 4}, 2)
	series.Medians()
	if series.Compared() == 0 {
		t.Fatal("比较次数应该被统计")
	}
}

// TestAgainstNaiveRandom 用朴素参考实现对随机数据逐窗口比对，覆盖奇/偶
// 窗口、窗口为 1、窗口等于数据、重复值等情况，共 1 万组。
func TestAgainstNaiveRandom(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	for group := 0; group < 10000; group++ {
		n := 1 + rng.Intn(30)
		var values []int64
		switch group % 4 {
		case 0:
			values = make([]int64, n)
			for i := range values {
				values[i] = rng.Int63n(20) - 10 // 大量重复值
			}
		case 1:
			values = make([]int64, n)
			for i := range values {
				values[i] = rng.Int63()
			}
		default:
			values = generate(n)
		}

		var window int
		switch group % 6 {
		case 0:
			window = 1
		case 1:
			window = n
		default:
			window = 1 + rng.Intn(n)
		}

		series := New(values, window)
		got := series.Medians()
		var naiveCompared int
		want := NaiveMedians(values, window, &naiveCompared)

		if len(got) != len(want) {
			t.Fatalf("第 %d 组窗口数不对：%d vs %d（n=%d w=%d）", group, len(got), len(want), n, window)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("第 %d 组位置 %d 中位数不一致：%v vs %v（n=%d w=%d values=%v）",
					group, i, got[i], want[i], n, window, values)
			}
		}
	}
}

// TestComparedOrderOfMagnitude 对照 README 的样例规模：增量实现的比较
// 次数应是 O(n log w)，而朴素实现约是 O(n·w·log w)。
func TestComparedOrderOfMagnitude(t *testing.T) {
	values := generate(20000)
	const window = 101

	series := New(values, window)
	series.Medians()
	incremental := series.Compared()

	var naive int
	NaiveMedians(values, window, &naive)

	bits := 0
	for bound := 1; bound < window; bound <<= 1 {
		bits++
	}
	limit := len(values) * bits
	if incremental > limit {
		t.Fatalf("增量比较次数应为 O(n log w) ≈ %d，实际 %d", limit, incremental)
	}
	if incremental*10 >= naive {
		t.Fatalf("增量比较次数 %d 应比朴素排序 %d 小一个数量级以上", incremental, naive)
	}
	t.Logf("n=%d w=%d：incremental=%d naive=%d（旧实现约 2009900）", len(values), window, incremental, naive)
}

// TestLargeInputComplexity 100 万个数、窗口 1001：比较次数受 O(n log w)
// 约束。朴素实现在这个规模上太慢，只检查增量实现。
func TestLargeInputComplexity(t *testing.T) {
	values := generate(1_000_000)
	const window = 1001

	series := New(values, window)
	medians := series.Medians()
	if len(medians) != len(values)-window+1 {
		t.Fatalf("窗口数应为 %d，得到 %d", len(values)-window+1, len(medians))
	}

	bits := 0
	for bound := 1; bound < window; bound <<= 1 {
		bits++
	}
	limit := len(values) * bits
	if compared := series.Compared(); compared > limit {
		t.Fatalf("大输入比较次数应为 O(n log w) ≈ %d，实际 %d", limit, compared)
	} else {
		t.Logf("n=%d w=%d：windows=%d compared=%d", len(values), window, len(medians), compared)
	}

	// 抽查若干窗口，与朴素实现一致。
	var zero int
	sample := []int{0, 12345, 500000, len(medians) - 1}
	for _, start := range sample {
		want := NaiveMedians(values[start:start+window], window, &zero)[0]
		if medians[start] != want {
			t.Fatalf("抽查窗口 %d 不一致：%v vs %v", start, medians[start], want)
		}
	}
}

func BenchmarkMedians(b *testing.B) {
	values := generate(1_000_000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		New(values, 1001).Medians()
	}
}
