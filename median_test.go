package slidingmedian

import (
	"math/rand"
	"testing"
)

func TestOddWindow(t *testing.T) {
	series := New([]int64{1, 3, 2, 5, 4}, 3)
	medians := series.Medians()
	if len(medians) != 3 {
		t.Fatalf("窗口数不对：%v", medians)
	}
	if medians[0] != 2 {
		t.Fatalf("第一个窗口的中位数应为 2：%v", medians)
	}
}

func TestEvenWindowAveragesTwoMiddles(t *testing.T) {
	medians := New([]int64{1, 2, 3, 4}, 4).Medians()
	if len(medians) != 1 || medians[0] != 2.5 {
		t.Fatalf("1,2,3,4 的中位数应为 2.5：%v", medians)
	}
	medians = New([]int64{4, 1, 3, 2}, 4).Medians()
	if len(medians) != 1 || medians[0] != 2.5 {
		t.Fatalf("乱序 4,1,3,2 的中位数应为 2.5：%v", medians)
	}
	medians = New([]int64{1, 2, 3, 4, 5}, 2).Medians()
	want := []float64{1.5, 2.5, 3.5, 4.5}
	for i := range want {
		if medians[i] != want[i] {
			t.Fatalf("窗口 %d 应为 %v：%v", i, want[i], medians)
		}
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
	if got := len(New([]int64{1, 2, 3}, 3).Medians()); got != 1 {
		t.Fatalf("应只有一个窗口：%d", got)
	}
}

func TestDuplicates(t *testing.T) {
	medians := New([]int64{2, 2, 2, 2}, 3).Medians()
	if len(medians) != 2 || medians[0] != 2 || medians[1] != 2 {
		t.Fatalf("重复值中位数不对：%v", medians)
	}
	medians = New([]int64{5, 1, 5, 1, 5}, 4).Medians()
	want := []float64{3, 3}
	for i := range want {
		if medians[i] != want[i] {
			t.Fatalf("窗口 %d 应为 %v：%v", i, want[i], medians)
		}
	}
}

func TestDoesNotModifyInput(t *testing.T) {
	input := []int64{9, 4, 7, 1, 8, 2, 6, 3, 5}
	snapshot := append([]int64{}, input...)
	New(input, 4).Medians()
	for i := range input {
		if input[i] != snapshot[i] {
			t.Fatalf("输入被修改：%v -> %v", snapshot, input)
		}
	}
}

func TestDeterministic(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	data := make([]int64, 500)
	for i := range data {
		data[i] = int64(rng.Intn(100))
	}
	first := New(data, 17).Medians()
	second := New(data, 17).Medians()
	if len(first) != len(second) {
		t.Fatalf("两次结果长度不同：%d vs %d", len(first), len(second))
	}
	for i := range first {
		if first[i] != second[i] {
			t.Fatalf("第 %d 个窗口两次结果不同：%v vs %v", i, first[i], second[i])
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

// 对 1 万组随机数据逐窗口比对增量实现与朴素参考实现。
func TestAgainstNaiveRandom(t *testing.T) {
	rng := rand.New(rand.NewSource(20260928))
	for trial := 0; trial < 10000; trial++ {
		n := rng.Intn(40)
		data := make([]int64, n)
		for i := range data {
			data[i] = int64(rng.Intn(11) - 5) // 小值域，制造大量重复
		}
		w := rng.Intn(n+3) - 1 // 覆盖 w<=0 与 w>n
		got := New(data, w).Medians()
		want, _ := MediansNaive(data, w)
		if len(got) != len(want) {
			t.Fatalf("trial %d: n=%d w=%d 窗口数 %d != %d", trial, n, w, len(got), len(want))
		}
		for i := range got {
			if got[i] != want[i] {
				t.Fatalf("trial %d: n=%d w=%d 窗口 %d：%v != %v（数据 %v）", trial, n, w, i, got[i], want[i], data)
			}
		}
	}
}

// 比较次数应是 n log w 量级，远小于朴素实现的 n×w。
func TestComparedMagnitude(t *testing.T) {
	const n, w = 20000, 101
	state := uint64(20260927)
	data := make([]int64, 0, n)
	for i := 0; i < n; i++ {
		state = state*6364136223846793005 + 1442695040888963407
		data = append(data, int64(state>>55)%1000)
	}
	series := New(data, w)
	series.Medians()
	_, naiveCompared := MediansNaive(data, w)
	if naiveCompared != (n-w+1)*w {
		t.Fatalf("朴素实现比较次数应为 %d：%d", (n-w+1)*w, naiveCompared)
	}
	if series.Compared() >= naiveCompared/4 {
		t.Fatalf("增量实现比较次数应远小于朴素实现：fast=%d naive=%d", series.Compared(), naiveCompared)
	}
	t.Logf("fast=%d naive=%d", series.Compared(), naiveCompared)
}

// 大输入：100 万个数、窗口 1001，验证结果与抽样朴素一致且比较次数受控。
func TestLargeInput(t *testing.T) {
	if testing.Short() {
		t.Skip("短模式跳过")
	}
	const n, w = 1000000, 1001
	state := uint64(99)
	data := make([]int64, 0, n)
	for i := 0; i < n; i++ {
		state = state*6364136223846793005 + 1442695040888963407
		data = append(data, int64(state>>33))
	}
	series := New(data, w)
	medians := series.Medians()
	if len(medians) != n-w+1 {
		t.Fatalf("窗口数应为 %d：%d", n-w+1, len(medians))
	}
	// 抽样与朴素实现比对
	for _, start := range []int{0, n / 2, n - w} {
		chunk := data[start : start+w]
		want, _ := MediansNaive(chunk, w)
		if medians[start] != want[0] {
			t.Fatalf("窗口 %d：%v != %v", start, medians[start], want[0])
		}
	}
	if limit := 50 * n; series.Compared() >= limit {
		t.Fatalf("比较次数应远小于 n*w：%d >= %d", series.Compared(), limit)
	}
	t.Logf("n=%d w=%d compared=%d", n, w, series.Compared())
}
