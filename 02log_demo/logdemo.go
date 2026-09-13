// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"02log_demo/internal/config"
	"02log_demo/internal/handler"
	"02log_demo/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/logdemo.yaml", "the config file")

func main() {
	flag.Parse()

	// 从配置文件中加载 go-zero HTTP 服务配置。
	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 创建 HTTP Server；main 退出时释放 go-zero 的相关资源。
	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	// ServiceContext 在服务启动时只创建一次，所有请求共享其中的 Pool。
	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	// monitorDone 用于通知状态监控 goroutine 退出；关闭 channel 可以广播退出信号。
	monitorDone := make(chan struct{})

	// 单独启动监控 goroutine，定期观察 Pool 的运行数、空闲对象数和排队任务数。
	go func() {
		// 每 250ms 生成一次事件；goroutine 结束时停止 ticker，释放定时器资源。
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				fmt.Printf(
					"pool状态 running=%d idle=%d queue=%d\n",
					ctx.Pool.GetRunningWorkersNum(),
					ctx.Pool.GetIdleWorkerCount(),
					ctx.Pool.GetTaskQueueLen(),
				)
			case <-monitorDone:
				return
			}
		}
	}()

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)

	// Windows 下当前 go-zero 版本不会替我们处理 Ctrl+C 的优雅退出。
	// 因此把阻塞的 server.Start 放到 goroutine 中，让 main 可以等待退出信号。
	go server.Start()

	// 容量为 1，确保 main 尚未读取信号时也能暂存一次 Ctrl+C 通知。
	quit := make(chan os.Signal, 1)

	// 将 Windows 的 Ctrl+C（os.Interrupt）转发到 quit channel。
	signal.Notify(quit, os.Interrupt)

	// 阻塞 main，直到用户按下 Ctrl+C。
	<-quit

	// 关闭 channel，使监控 goroutine 的退出分支立即就绪。
	close(monitorDone)
	fmt.Println("收到退出信号，开始等待日志任务完成……")

	// Close 停止接收新任务，但不会丢弃已经提交的任务。
	ctx.Pool.Close()

	// Wait 等待正在执行和排队中的任务全部结束，避免日志随进程退出而丢失。
	ctx.Pool.Wait()
	fmt.Println("日志任务全部完成，程序退出")
}
