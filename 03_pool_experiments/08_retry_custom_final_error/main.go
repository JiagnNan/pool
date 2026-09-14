package main

import (
	"fmt"
	agilepool "github.com/Yiming1997/agilePool/v2"
	"time"
)

func main() {
	config := agilepool.NewConfig(
		// 只允许一个 Worker，让每次执行和退避等待按顺序发生，便于观察时间间隔。
		agilepool.WithWorkerNumCapacity(1),

		// 使用阻塞模式，避免本实验中的任务被立即拒绝。
		agilepool.WithBlockMode(agilepool.BLOCK),

		// 为任务预留足够的通道容量，排除队列饱和的影响。
		agilepool.WithTaskQueueSize(10),
	)

	pool := agilepool.NewPool(config)

	// attempts 记录首次执行和后续重试的总调用次数。
	attempts := 0

	// lastErr 由业务任务主动保存，用来证明最终错误不是由 Pool 或 Wait() 返回的。
	var lastErr error

	// 记录整个重试任务从提交到结束所经过的时间。
	start := time.Now()

	retryTask := &agilepool.TaskWithRetry{
		MinBackOff: 50 * time.Millisecond,
		MaxBackOff: 100 * time.Millisecond,
		RetryNum:   3,

		// 自定义策略故意返回超过 MaxBackOff 的时间，验证库不会再次限制返回值。
		BackOffStrategy: func(
			min time.Duration,
			max time.Duration,
			retryNum uint,
		) time.Duration {
			delay := time.Duration(retryNum) * (200 * time.Millisecond)

			fmt.Printf(
				"第 %d 次重试：min=%v max=%v 自定义等待=%v\n",
				retryNum,
				min,
				max,
				delay,
			)

			// 这里没有使用 max 进行截断，所以返回多少，协程池就等待多久。
			return delay
		},

		// 本实验让首次执行和三次重试全部失败。
		Task: func() error {
			attempts++

			lastErr = fmt.Errorf("第 %d 次执行仍然失败", attempts)

			fmt.Println(
				time.Now().Format("15:04:05.000"),
				"任务执行，次数：",
				attempts,
			)

			// fmt.Errorf() 只创建错误；这里只返回错误，不主动打印错误内容。
			return lastErr
		},
	}

	pool.Submit(retryTask)

	// Close() 拒绝后续新任务，Wait() 等待首次执行和三次重试全部结束。
	pool.Close()
	pool.Wait()

	// 最终错误之所以能够显示，是因为业务代码通过 lastErr 主动保存并打印了它。
	fmt.Println("实际执行次数：", attempts)
	fmt.Println("业务代码自行保存的最后错误：", lastErr)
	fmt.Println("总耗时：", time.Since(start))
}
