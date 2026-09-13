package main

import (
	"fmt"
	"time"

	agilepool "github.com/Yiming1997/agilePool/v2"
)

func main() {
	// 从实验开始处计时，用于观察 Wait 返回前任务实际运行了多久。
	start := time.Now()

	config := agilepool.NewConfig(
		// 将 Worker 最大数量限制为 1，方便观察单个任务的完整执行过程。
		agilepool.WithWorkerNumCapacity(1),

		// 为快速任务通道设置足够容量，避免本实验受到队列容量影响。
		agilepool.WithTaskQueueSize(10),

		// 使用阻塞模式，确保实验重点只放在 SubmitBefore 的开始期限上。
		agilepool.WithBlockMode(agilepool.BLOCK),
	)

	pool := agilepool.NewPool(config)

	// 任务 C 能在 1 秒期限内开始，但自身需要运行 3 秒。
	pool.SubmitBefore(
		agilepool.TaskFunc(func() error {
			fmt.Println(
				time.Now().Format("15:04:05.000"),
				"任务C开始",
			)

			// 运行时间超过开始期限，用来证明任务开始后不会被强制终止。
			time.Sleep(3 * time.Second)

			fmt.Println(
				time.Now().Format("15:04:05.000"),
				"任务C结束",
			)

			return nil
		}),

		// 该期限只约束任务何时开始，不约束任务开始后的运行时长。
		1*time.Second,
	)

	// 拒绝后续新任务，但保留已经提交的任务。
	pool.Close()

	// 等待任务 C 完整运行结束。
	pool.Wait()

	fmt.Println(
		"总耗时：",
		time.Since(start),
	)
}
