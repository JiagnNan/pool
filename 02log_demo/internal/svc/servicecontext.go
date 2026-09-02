// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"02log_demo/internal/config"
	agilepool "github.com/Yiming1997/agilePool/v2"
)

type ServiceContext struct {
	Config config.Config
	Pool   *agilepool.Pool
}

func NewServiceContext(c config.Config) *ServiceContext {
	pool := agilepool.NewPool(agilepool.NewConfig(agilepool.WithWorkerNumCapacity(3)))
	return &ServiceContext{
		Config: c,
		Pool:   pool,
	}
}
