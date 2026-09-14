package main

import (
	"fmt"
	agilepool "github.com/Yiming1997/agilePool/v2"
)

func main() {
	config := agilepool.NewConfig(
		// 只允许一个 Worker，便于确认发生 panic 后是否由同一条消费链继续处理任务 B。
		agilepool.WithWorkerNumCapacity(1),

		// 为两个任务预留足够的通道容量，排除队列饱和对本实验的影响。
		agilepool.WithTaskQueueSize(10),

		// 使用阻塞模式，保证本实验只观察 panic 恢复与任务计数平衡。
		agilepool.WithBlockMode(agilepool.BLOCK),
	)

	pool := agilepool.NewPool(config)

	// 任务 A 主动触发 panic，由 worker.runTask() 内部的 recover 负责捕获。
	taskA := agilepool.TaskFunc(func() error {
		fmt.Println("任务A开始")
		panic("模拟任务发生 panic")
	})

	// 任务 B 排在任务 A 后面，用于验证一次任务 panic 不会阻止后续任务执行。
	taskB := agilepool.TaskFunc(func() error {
		fmt.Println("任务B开始")
		fmt.Println("任务B结束")
		return nil
	})

	// 按顺序提交两个任务；任务 A 的 panic 不会由提交任务的主 goroutine 直接触发。
	pool.Submit(taskA)
	pool.Submit(taskB)

	// 停止接收新任务，并等待两个已接收任务的计数全部归零。
	pool.Close()
	pool.Wait()

	// 能运行到这里，说明 panic 已恢复，并且两个任务的 WaitGroup 计数都得到平衡。
	fmt.Println("所有任务处理结束")
}
