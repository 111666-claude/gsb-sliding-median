package slidingmedian

import "testing"

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

func TestWindowEqualsData(t *testing.T) {
if got := len(New([]int64{1, 2, 3}, 3).Medians()); got != 1 {
t.Fatalf("应只有一个窗口：%d", got)
}
}

func TestComparedIsCounted(t *testing.T) {
series := New([]int64{1, 2, 3, 4}, 2)
series.Medians()
if series.Compared() == 0 {
t.Fatal("比较次数应该被统计")
}
}
