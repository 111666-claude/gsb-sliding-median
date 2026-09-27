// Command sliding-median 计算滑动窗口中位数。
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"example.com/slidingmedian"
)

// values 生成固定种子的数据。
func values(count int) []int64 {
	out := make([]int64, 0, count)
	state := uint64(20260927)
	for index := 0; index < count; index++ {
		state = state*6364136223846793005 + 1442695040888963407
		out = append(out, int64(state>>55)%1000)
	}
	return out
}

// Run 执行一次命令行调用，返回退出码。
func Run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("sliding-median", flag.ContinueOnError)
	flags.SetOutput(stderr)
	raw := flags.String("values", "", "逗号分隔的整数序列")
	generate := flags.Int("generate", 0, "生成这么多个数（与 --window 一起用）")
	window := flags.Int("window", 4, "窗口大小")
	counters := flags.Bool("counters", false, "打印比较次数")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	var data []int64
	if *raw != "" {
		for _, field := range strings.Split(*raw, ",") {
			field = strings.TrimSpace(field)
			if field == "" {
				continue
			}
			var value int64
			if _, err := fmt.Sscan(field, &value); err != nil {
				fmt.Fprintln(stderr, "不是整数")
				return 2
			}
			data = append(data, value)
		}
	} else {
		data = values(*generate)
	}
	series := slidingmedian.New(data, *window)
	medians := series.Medians()
	if len(medians) <= 8 {
		fmt.Fprintf(stdout, "medians=%v\n", medians)
	}
	fmt.Fprintf(stdout, "windows=%d\n", len(medians))
	if *counters {
		fmt.Fprintf(stdout, "compared=%d\n", series.Compared())
	}
	return 0
}

func main() {
	os.Exit(Run(os.Args[1:], os.Stdout, os.Stderr))
}
