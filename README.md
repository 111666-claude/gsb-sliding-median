# sliding-median

滑动窗口中位数，只用 Go 标准库。

```
go test ./...
go run ./cmd/sliding-median --values "1,2,3,4" --window 4
go run ./cmd/sliding-median --generate 20000 --window 101 --counters
```

## 口径（README 为准）

- **偶数窗口取两个中间值的平均**：`--values "1,2,3,4" --window 4` 按口径是 `2.5`，现在给 `3`（取了下中位）。
- **增量维护**：窗口滑动时用两个堆（或有序容器）做 O(log w) 更新，整体 O(n log w)。
  `--generate 20000 --window 101 --counters` 现在 `compared=2009900`（每个窗口都重排），
  按口径应该在同一量级（n log w，约 14 万）。
- 中位数返回 `float64`：奇数窗口是中间值，偶数窗口是两中值平均；窗口大于数据或窗口 ≤ 0 返回空。
- 不变量：窗口数等于 `n - w + 1`；每个窗口的中位数与该窗口排序后的中位数一致；
  同一输入重复计算结果相同。
- 规模：100 万个数的数据、窗口 1001 时内存 O(w)，时间 O(n log w)；不许每个窗口重排。

## 现在的行为

- `Medians` 用两个堆（下半大顶堆、上半小顶堆）加延迟删除增量维护：
  每次滑动 O(log w)，整体 O(n log w)，内存 O(w)，不复制窗口内容；
  `--generate 20000 --window 101 --counters` 的 `compared` 从 2009900 降到约 35 万。
- 偶数窗口取两个中间值的平均：`1,2,3,4` 给 `2.5`。
- 窗口大于数据、window ≤ 0 或数据为空时返回空。
- `MediansNaive`（naive.go）是朴素参考实现（每个窗口复制 + 排序），
  测试用它对 1 万组随机数据逐窗口对照，并比较比较次数量级。

## 计划新增的能力

```
--values 列表    逗号分隔的数据
--generate N     生成 N 个固定种子的数据
--counters       打印比较次数
```

## 输出

```
medians=[…]
windows=… compared=…
```

## 目录

```
median.go              滑动中位数
cmd/sliding-median     命令行入口
median_test.go         go test 用例
```
