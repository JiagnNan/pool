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

	// 任务 C 主动监听 context：正常情况下等待两秒，收到取消信号时提前退出。
	taskC := agilepool.TaskFunc(func() error {
		fmt.Println(time.Now().Format("15:04:05.000"), "任务C开始")

		// 关闭通道，通知主 goroutine。
		close(started)

		// 同时等待“正常工作完成”和“context 被取消”，先发生的分支先执行。
		select {
		case <-time.After(2 * time.Second):
			fmt.Println("任务C正常完成")

		case <-ctx.Done():
			fmt.Println("任务C收到取消信号：", ctx.Err())
			// 提前返回后，不会执行 select 下方的“任务C结束”。
			return ctx.Err()
		}

		// 只有两秒计时先完成时，代码才会运行到这里。
		fmt.Println(time.Now().Format("15:04:05.000"), "任务C结束")
		return nil
	})

	// SubmitCtx 会用 contextTask 包装任务 C，再交给协程池。
	pool.SubmitCtx(ctx, taskC)

	// 确认任务 C 已经通过 contextTask.Process() 的取消检查并开始执行。
	<-started

	fmt.Println("主 goroutine 取消 context")
	// 协程池不会再次检查取消信号，但任务 C 自己正在 select 中监听它。
	cancel()

	// 停止接收新任务，并等待任务 C 正常完成或主动取消后退出。
	pool.Close()
	pool.Wait()
}
