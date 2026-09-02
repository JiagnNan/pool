package main

import (
	"fmt"
	"sync"
	"time"
)

func printGoroutine(v int, wg *sync.WaitGroup, ch chan struct{}) {
	defer func() {
		<-ch
		wg.Done()
	}()
	println("task", v, "start")
	time.Sleep(time.Second)
	println("task", v, "end")
}

//func main() {
//	var wg sync.WaitGroup
//	ch := make(chan struct{}, 3)
//	for i := 0; i < 20; i++ {
//		wg.Add(1)
//		ch <- struct{}{}
//		go printGoroutine(i, &wg, ch)
//	}
//	wg.Wait()
//}

func worker(id int, jobs <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		fmt.Println("worker", id, "start task", job)
		time.Sleep(time.Second)
		fmt.Println("worker", id, "end task", job)
	}
}
func main() {
	var wg sync.WaitGroup
	ch := make(chan int, 5)
	wg.Add(3)
	for i := 0; i < 3; i++ {
		go worker(i, ch, &wg)
	}
	//发送20个任务
	for i := 0; i < 20; i++ {
		ch <- i
	}

	close(ch)

	wg.Wait()
}
