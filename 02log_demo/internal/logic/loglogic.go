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

	"github.com/zeromicro/go-zero/core/logx"

	agilepool "github.com/Yiming1997/agilePool/v2"
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

	id := req.Id
	action := req.Action

	// TaskFunc 把 func() error 包装为 agilePool 任务。
	task := agilepool.TaskFunc(func() error {
		// 获取 Pool 当前正在运行的 Worker 数量。
		// 这是公开 API，不涉及阅读内部源码。
		running := l.svcCtx.Pool.GetRunningWorkersNum()

		// 记录任务真正开始执行的时间。
		// 精确到毫秒，方便观察哪些任务同时开始。
		fmt.Printf("%s task id=%d start runningWorkers=%d\n", time.Now().Format("15:04:05.000"), id, running)

		// 无论任务成功还是返回错误，结束前都打印 end。
		defer func() {
			fmt.Printf("%s task id=%d end\n", time.Now().Format("15:04:05.000"), id)
		}()

		time.Sleep(time.Second * 1)
		file, err := os.OpenFile(
			"logs/business.log",
			os.O_CREATE|os.O_WRONLY|os.O_APPEND,
			0644,
		)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = fmt.Fprintf(
			file, //写入到哪个文件
			"%s id=%d action=%s\n",
			time.Now().Format("2006-01-02 15:04:05"),
			id,
			action,
		)
		if err != nil {
			return err
		}
		return nil
	})

	l.svcCtx.Pool.Submit(task)

	res := &types.LogResp{
		Message: action,
	}

	return res, nil
}
