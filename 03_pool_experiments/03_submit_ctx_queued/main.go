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

	// started 只传递“任务 A 已经开始”的信号，不传递具体数据。
	started := make(chan struct{})

	// 任务 A 占住唯一的 Worker，使任务 B 只能进入队列等待。
	taskA := agilepool.TaskFunc(func() error {
		fmt.Println(time.Now().Format("15:04:05.000"), "任务A开始")

		// 关闭通道，通知主 goroutine。
		close(started)

		time.Sleep(2 * time.Second)
		fmt.Println(time.Now().Format("15:04:05.000"), "任务A结束")
		return nil
	})

	// 如果排队期间 context 已经过期，任务 B 的函数体不会被执行。
	taskB := agilepool.TaskFunc(func() error {
		fmt.Println(time.Now().Format("15:04:05.000"), "任务B开始")
		return nil
	})

	pool.Submit(taskA)

	// 等到任务 A 确实开始后，再为任务 B 创建带超时的 context。
	<-started

	ctx, cancel := context.WithTimeout(
		context.Background(),
		500*time.Millisecond,
	)
	defer cancel()

	// context 从这里提交时开始计时；任务 B 最多只有 500 毫秒的等待机会。
	pool.SubmitCtx(ctx, taskB)

	// 停止接收新任务，并等待已经接收的包装任务全部处理完成。
	pool.Close()
	pool.Wait()
}
