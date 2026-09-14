package main

import (
	"fmt"
	agilepool "github.com/Yiming1997/agilePool/v2"
)

func main() {
	config := agilepool.NewConfig(
		// 只允许一个 Worker，便于确定任务 A 正在执行时没有其他 Worker 消费编号任务。
		agilepool.WithWorkerNumCapacity(1),

		// 使用阻塞模式，避免本实验中的任务被立即拒绝。
		agilepool.WithBlockMode(agilepool.BLOCK),

		// 将 taskQueue 容量设为 1，使任务 1 留在通道中，其余任务进入 taskBuf。
		agilepool.WithTaskQueueSize(1),
	)

	pool := agilepool.NewPool(config)

	started := make(chan struct{})
	release := make(chan struct{})

	executionOrder := make([]int, 0, 11)
	acceptedCount := 0

	taskA := agilepool.TaskFunc(func() error {
		fmt.Println("任务A开始，占用唯一的 Worker")

		// 只由任务 A 关闭一次。
		close(started)

		// 等待主 goroutine 释放。
		<-release

		fmt.Println("任务A结束")
		return nil
	})

	pool.Submit(taskA)

	// 确认任务 A 已经占住 Worker，再提交编号任务。
	<-started

	for i := 1; i <= 11; i++ {
		// 明确保存本轮任务编号，供闭包使用。
		taskID := i

		task := agilepool.TaskFunc(func() error {
			fmt.Println("执行任务：", taskID)
			executionOrder = append(executionOrder, taskID)
			return nil
		})

		// BLOCK 模式下，taskQueue 已满时任务会尝试进入 taskBuf。
		if pool.TrySubmit(task) {
			acceptedCount++
		}
	}
	fmt.Println("成功接收任务数量：", acceptedCount)
	// 这里只统计 taskQueue，不包含 taskBuf 中的任务数量。
	fmt.Println("taskQueue当前长度：", pool.GetTaskQueueLen())

	// 释放任务 A，让唯一的 Worker 开始处理排队任务。
	close(release)

	pool.Close()
	pool.Wait()

	fmt.Println("最终执行顺序：", executionOrder)
}
