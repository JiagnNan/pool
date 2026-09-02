// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"

	"02log_demo/internal/config"
	"02log_demo/internal/handler"
	"02log_demo/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/logdemo.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	go server.Start()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, os.Interrupt)

	<-quit
	fmt.Println("收到退出信号，开始等待日志任务完成……")

	ctx.Pool.Close()
	ctx.Pool.Wait()
	fmt.Println("日志任务全部完成，程序退出")
}
