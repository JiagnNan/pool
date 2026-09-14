package main

import (
	"fmt"
	agilepool "github.com/Yiming1997/agilePool/v2"
	"time"
)

func main() {
	config := agilepool.NewConfig(
		// 只允许一个 Worker，用来观察重试期间是否会执行后面的普通任务。
		agilepool.WithWorkerNumCapacity(1),

		// 使用阻塞模式，避免任务因为通道暂时繁忙而被拒绝。
		agilepool.WithBlockMode(agilepool.BLOCK),

		// 为两个任务预留足够的通道容量，排除队列饱和的影响。
		agilepool.WithTaskQueueSize(10),
	)

	pool := agilepool.NewPool(config)

	// attempts 记录第一次执行和后续重试的总调用次数。
	attempts := 0

	// 前三次执行返回错误，第四次执行成功。
	retryTask := &agilepool.TaskWithRetry{
		MinBackOff: 100 * time.Millisecond,
		MaxBackOff: 500 * time.Millisecond,
		RetryNum:   3,
		Task: func() error {
			attempts++

			fmt.Printf(
				"%s 重试任务执行，次数：%d\n",
				time.Now().Format("15:04:05.000"),
				attempts,
			)

			if attempts < 4 {
				return fmt.Errorf("第 %d 次模拟失败", attempts)
			}

			fmt.Println(time.Now().Format("15:04:05.000"), "重试任务执行成功")
			return nil
		},
	}

	// 普通任务 B 排在重试任务后面，用于证明退避 Sleep 会继续占用 Worker。
	taskB := agilepool.TaskFunc(func() error {
		fmt.Println(
			time.Now().Format("15:04:05.000"),
			"普通任务B开始",
		)
		return nil
	})

	start := time.Now()

	// 同一个 Worker 会先完成整个重试流程，再执行任务 B。
	pool.Submit(retryTask)
	pool.Submit(taskB)

	// 停止接收新任务，并等待重试任务和普通任务 B 全部完成。
	pool.Close()
	pool.Wait()

	// 通过次数与总耗时验证 RetryNum 和默认指数退避策略。
	fmt.Println("实际尝试次数：", attempts)
	fmt.Println("总耗时：", time.Since(start))
}
