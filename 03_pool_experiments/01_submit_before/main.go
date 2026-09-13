package main

import (
	"fmt"
	"time"

	agilepool "github.com/Yiming1997/agilePool/v2"
)

func main() {
	config := agilepool.NewConfig(
		// 将 Worker 最大数量限制为 1，确保任务只能依次执行。
		agilepool.WithWorkerNumCapacity(1),

		// 为快速任务通道设置足够的容量，使任务 B 能被接收并排队。
		agilepool.WithTaskQueueSize(10),

		// 使用阻塞模式，任务通道繁忙时允许提交方等待或进入溢出缓冲区。
		agilepool.WithBlockMode(agilepool.BLOCK),
	)

	pool := agilepool.NewPool(config)

	// 任务A会占用唯一的Worker。
	pool.Submit(
		agilepool.TaskFunc(func() error {
			fmt.Println(
				time.Now().Format("15:04:05.000"),
				"任务A开始",
			)

			// 休眠 2 秒，模拟长期占用唯一 Worker 的耗时任务。
			time.Sleep(2 * time.Second)

			fmt.Println(
				time.Now().Format("15:04:05.000"),
				"任务A结束",
			)

			return nil
		}),
	)

	// 任务B的开始期限只有500毫秒。
	// 因为唯一的Worker被任务A占用，所以任务B应该在执行前超时。
	pool.SubmitBefore(
		agilepool.TaskFunc(func() error {
			fmt.Println(
				time.Now().Format("15:04:05.000"),
				"任务B执行了",
			)

			return nil
		}),

		// 任务 B 必须在提交后的 500 毫秒内开始，否则跳过原任务。
		time.Millisecond*500,
	)

	// 不再接受新任务，但允许已经提交的任务完成。
	pool.Close()

	// 等待任务A和任务B的包装任务全部处理完成。
	pool.Wait()

	fmt.Println(
		time.Now().Format("15:04:05.000"),
		"所有任务处理结束",
	)
}
