package main

import (
	"context"
	"fmt"
	agilepool "github.com/Yiming1997/agilePool/v2"
	"time"
)

func main() {
	config := agilepool.NewConfig(
		// 将 Worker 最大数量限制为 1，方便观察单个任务的完整执行过程。
		agilepool.WithWorkerNumCapacity(1),

		// 使用阻塞模式，确保实验重点只放在 SubmitCtx 的取消行为上。
		agilepool.WithBlockMode(agilepool.BLOCK),
	)

	pool := agilepool.NewPool(config)

	// 使用可主动取消的 context，方便在任务 C 开始后精确发送取消信号。
	ctx, cancel := context.WithCancel(context.Background())

	// started 用于通知主 goroutine：任务 C 已经进入函数体。
	started := make(chan struct{})

	// 任务 C 开始后持续运行两秒，用来观察取消 context 是否会强制终止它。
	taskC := agilepool.TaskFunc(func() error {
		fmt.Println(time.Now().Format("15:04:05.000"), "任务C开始")

		// 关闭通道，通知主 goroutine。
		close(started)

		time.Sleep(2 * time.Second)
		fmt.Println(time.Now().Format("15:04:05.000"), "任务C结束")
		return nil
	})

	// SubmitCtx 会用 contextTask 包装任务 C，再交给协程池。
	pool.SubmitCtx(ctx, taskC)

	// 确认任务 C 已经通过 contextTask.Process() 的取消检查并开始执行。
	<-started

	fmt.Println("主 goroutine 取消 context")
	// 此时任务 C 已经开始，取消信号不会被协程池再次检查。
	cancel()

	// 停止接收新任务，并等待任务 C 完整结束。
	pool.Close()
	pool.Wait()
}
