package main

import (
	"fmt"
	"time"

	agilepool "github.com/Yiming1997/agilePool/v2"
)

func main() {
	config := agilepool.NewConfig(
		// 只允许一个 Worker，使任务 A 运行时任务 B 只能在队列中等待。
		agilepool.WithWorkerNumCapacity(1),

		// 使用 BLOCK 模式，让关闭前提交的任务按照阻塞模式的规则进入协程池。
		agilepool.WithBlockMode(agilepool.BLOCK),

		// taskQueue 只需容纳等待中的任务 B。
		agilepool.WithTaskQueueSize(1),
	)

	pool := agilepool.NewPool(config)

	// started 用于确认任务 A 已经占用唯一的 Worker。
	started := make(chan struct{})

	// release 用于通知任务 A 可以结束。
	release := make(chan struct{})

	// waitStarted 用于确认负责调用 Wait() 的 goroutine 已经开始运行。
	waitStarted := make(chan struct{})

	// waitReturned 只会在 Wait() 真正返回后关闭。
	waitReturned := make(chan struct{})

	// 任务 A 会一直占用唯一的 Worker，直到主 goroutine 关闭 release。
	taskA := agilepool.TaskFunc(func() error {
		fmt.Println("任务A开始")
		close(started)
		<-release
		fmt.Println("任务A结束")
		return nil
	})

	// 任务 B 在 Close() 前提交，用于验证已接收任务在关闭后仍会执行。
	taskB := agilepool.TaskFunc(func() error {
		fmt.Println("任务B开始")
		fmt.Println("任务B结束")
		return nil
	})

	// 任务 C 在 Close() 后通过 TrySubmit() 提交，正常情况下不会执行。
	taskC := agilepool.TaskFunc(func() error {
		fmt.Println("任务C开始")
		fmt.Println("任务C结束")
		return nil
	})

	// 任务 D 在 Close() 后通过没有返回值的 Submit() 提交，也不应执行。
	taskD := agilepool.TaskFunc(func() error {
		fmt.Println("任务D开始")
		fmt.Println("任务D结束")
		return nil
	})

	// 先启动任务 A，并等待它确实占住 Worker。
	pool.Submit(taskA)
	<-started

	// 此时任务 B 会进入 taskQueue，返回 true 只代表已经被协程池接收。
	acceptedBeforeClose := pool.TrySubmit(taskB)
	fmt.Println("关闭前任务B是否被接收：", acceptedBeforeClose)

	// 第一次调用将 Pool 标记为关闭；第二次调用用于验证 Close() 的幂等性。
	pool.Close()
	pool.Close()

	// TrySubmit() 可以通过返回值明确反映关闭后的拒绝结果。
	acceptedAfterClose := pool.TrySubmit(taskC)
	fmt.Println("关闭后任务C是否被接收：", acceptedAfterClose)

	// Submit() 没有返回值，只能通过任务 D 没有打印日志来判断它未被执行。
	pool.Submit(taskD)

	go func() {
		// 先通知主 goroutine：本 goroutine 已经运行到调用 Wait() 之前。
		close(waitStarted)
		pool.Wait()

		// 只有 Wait() 真正返回后，才关闭该信号通道。
		close(waitReturned)
	}()

	// 确保 Wait goroutine 已经运行，再开始计时判断。
	<-waitStarted

	// 任务 A 尚未释放且任务 B 尚未执行，因此 Wait() 此时不应返回。
	select {
	case <-waitReturned:
		fmt.Println("释放任务A之前，Wait已经返回")
	case <-time.After(300 * time.Millisecond):
		fmt.Println("释放任务A之前，Wait仍在等待")
	}

	// 释放任务 A，随后唯一的 Worker 会继续执行已经接收的任务 B。
	close(release)

	// 等待任务 A、B 全部结束以及 pool.Wait() 真正返回。
	<-waitReturned

	fmt.Println("所有已接收任务执行完成")
}
