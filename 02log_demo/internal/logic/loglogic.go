// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"fmt"
	"os"
	"time"

	"02log_demo/internal/svc"
	"02log_demo/internal/types"

	agilepool "github.com/Yiming1997/agilePool/v2"
	"github.com/zeromicro/go-zero/core/logx"
)

type LogLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LogLogic {
	return &LogLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LogLogic) Log(req *types.LogReq) (resp *types.LogResp, err error) {
	// 在 HTTP Handler 返回前复制任务所需的数据。
	// 后台任务可能稍后才开始，因此不继续依赖请求对象 req。
	id := req.Id
	action := req.Action

	// TaskFunc 把 func() error 形式的业务函数包装成 agilePool 可以执行的任务。
	task := agilepool.TaskFunc(func() error {
		// 获取 Pool 当前正在运行的 Worker 数量。
		// 该数据只用于本次学习实验，帮助观察实际并发数量。
		running := l.svcCtx.Pool.GetRunningWorkersNum()

		// 记录任务真正开始执行的时间。
		// 精确到毫秒，方便观察哪些任务同时开始。
		fmt.Printf("%s task id=%d start runningWorkers=%d\n", time.Now().Format("15:04:05.000"), id, running)

		// 无论任务成功还是返回错误，结束前都打印 end。
		defer func() {
			fmt.Printf("%s task id=%d end\n", time.Now().Format("15:04:05.000"), id)
		}()

		// 仅在指定测试请求中主动触发 panic，用来验证 Worker 的异常隔离能力。
		// agilePool 会在 runTask 中 recover，因此当前任务会中止，但服务和后续任务仍可继续运行。
		if action == "panic_test" {
			panic("模拟日志任务 panic")
		}

		// 本次实验用 200ms 模拟耗时操作，让监控器能够观察 Worker 从运行到空闲、再到过期清理的变化。
		time.Sleep(200 * time.Millisecond)

		// 以“创建、只写、追加”模式打开日志文件，避免覆盖之前的日志。
		file, err := os.OpenFile(
			"logs/business.log",
			os.O_CREATE|os.O_WRONLY|os.O_APPEND,
			0644,
		)
		if err != nil {
			// HTTP 响应可能已经返回，无法再把后台错误交给客户端。
			// 因此在任务内部记录任务数据和底层文件错误，避免静默失败。
			logx.Errorf(
				"异步日志打开文件失败: id=%d action=%s error=%v",
				id,
				action,
				err,
			)
			return err
		}

		// 任务退出时关闭文件句柄，避免长期运行后耗尽系统文件资源。
		defer file.Close()

		// 将本次业务日志格式化后写入文件，并使用换行分隔不同任务。
		_, err = fmt.Fprintf(
			file,
			"%s id=%d action=%s\n",
			time.Now().Format("2006-01-02 15:04:05"),
			id,
			action,
		)
		if err != nil {
			// 文件已成功打开仍不代表写入一定成功，写入错误也要单独记录。
			logx.Errorf(
				"异步日志写入文件失败: id=%d action=%s error=%v",
				id,
				action,
				err,
			)
			return err
		}
		return nil
	})

	// TrySubmit 会返回任务是否被协程池接受，避免像 Submit 一样忽略提交结果。
	accepted := l.svcCtx.Pool.TrySubmit(task)

	// false 表示任务既没有开始执行，也没有进入等待队列，需要由业务自行处理。
	if !accepted {
		logx.Errorf("异步日志任务提交失败: id=%d action=%s", id, action)

		return &types.LogResp{
			Message: "日志任务提交失败",
		}, nil
	}

	// 这里的响应只表示日志任务已被接收，不表示日志已经写入文件。
	res := &types.LogResp{
		Message: action,
	}

	return res, nil
}
