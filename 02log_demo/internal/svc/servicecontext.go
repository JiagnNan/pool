// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"02log_demo/internal/config"
	agilepool "github.com/Yiming1997/agilePool/v2"
)

type ServiceContext struct {
	Config config.Config

	// Pool 是整个服务共享的协程池，所有 LogLogic 都向同一个 Pool 提交任务。
	Pool *agilepool.Pool
}

func NewServiceContext(c config.Config) *ServiceContext {
	// Pool 只在服务启动时创建一次。
	pool := agilepool.NewPool(
		agilepool.NewConfig(
			// 限制同时运行的 Worker 数量；它是最大值，不是启动时固定创建的数量。
			agilepool.WithWorkerNumCapacity(3),

			// taskQueue 最多暂存 100 个等待 Worker 接收的任务，为突发任务提供排队空间。
			agilepool.WithTaskQueueSize(100),

			// BLOCK 模式会优先排队而不是立即拒绝任务，便于观察调度器能否根据积压扩容。
			agilepool.WithBlockMode(agilepool.BLOCK),

			// 本次实验只保留最近一次速率采样，使调度器能快速感知短时突发流量。
			// 窗口越小越灵敏，但也更容易受到瞬时波动影响，不代表生产环境的通用配置。
			agilepool.WithStatsWindowSize(1),
		),
	)

	return &ServiceContext{
		Config: c,
		Pool:   pool,
	}
}
