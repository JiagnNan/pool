package main

import (
	"fmt"
	agilepool "github.com/Yiming1997/agilePool/v2"
)

func main() {
	// 先验证任务通道满时立即拒绝的 NONBLOCK 模式。
	runExperiment(
		"NONBLOCK",
		agilepool.NONBLOCK,
	)

	fmt.Println()

	// 再使用完全相同的任务验证 BLOCK 模式下的分块缓冲行为。
	runExperiment(
		"BLOCK",
		agilepool.BLOCK,
	)
}

func runExperiment(
	modeName string,
	mode agilepool.WorkMode,
) {
	fmt.Println("开始实验模式：", modeName)

	config := agilepool.NewConfig(
		// 唯一的 Worker 被任务 A 占用后，不能同时消费任务 B、C。
		agilepool.WithWorkerNumCapacity(1),

		// 通道只能保存一个等待任务，任务 B 入队后通道立即达到容量上限。
		agilepool.WithTaskQueueSize(1),

		// 同一段实验代码通过参数分别使用 NONBLOCK 和 BLOCK。
		agilepool.WithBlockMode(mode),
	)

	pool := agilepool.NewPool(config)

	// 任务 A 通过它通知主 goroutine：我已经开始。
	started := make(chan struct{})

	// 主 goroutine 通过它通知任务 A：现在可以结束。
	release := make(chan struct{})

	taskA := agilepool.TaskFunc(func() error {
		fmt.Println("任务A开始，占用唯一的 Worker")

		close(started)

		// 一直等待主 goroutine 关闭 release。
		<-release

		fmt.Println("任务A结束")
		return nil
	})

	// 任务 B 用于占满容量为 1 的 taskQueue。
	taskB := agilepool.TaskFunc(func() error {
		fmt.Println("任务B开始 Worker")
		fmt.Println("任务B结束")
		return nil
	})

	// 任务 C 用于观察队列饱和后，是被拒绝还是进入 taskBuf。
	taskC := agilepool.TaskFunc(func() error {
		fmt.Println("任务C开始 Worker")
		fmt.Println("任务C结束")
		return nil
	})

	// 先启动任务 A，并让它阻塞在 release 上。
	pool.Submit(taskA)

	// 确保任务 A 已经占住唯一的 Worker。
	<-started

	// 任务 B 应该进入容量为 1 的 taskQueue。
	acceptedB := pool.TrySubmit(taskB)

	// 此时 Worker 被 A 占用，taskQueue 又被 B 占满：
	// NONBLOCK 会拒绝 C，BLOCK 会把 C 保存进 taskBuf。
	acceptedC := pool.TrySubmit(taskC)

	fmt.Println("任务B是否被接收：", acceptedB)
	fmt.Println("任务C是否被接收：", acceptedC)

	// 释放任务 A，否则 Wait() 会一直等待。
	close(release)

	// 停止接收新任务，并等待所有已经成功接收的任务处理完毕。
	pool.Close()
	pool.Wait()
}
