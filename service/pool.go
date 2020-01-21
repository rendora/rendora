package service

import (
	"time"

	"github.com/silenceper/pool"
)

// NewHeadlessClientPool creates HeadlessClientPool
func NewHeadlessClientPool(cfg *HeadlessConfig) (pool.Pool, error) {
	factory := func() (interface{}, error) {
		headlessClient, err := NewHeadlessClient(cfg)
		if err != nil {
			return nil, err
		}

		return headlessClient, nil
	}

	close := func(v interface{}) error {
		return v.(*HeadlessClient).Close()
	}

	ping := func(v interface{}) error {
		return v.(*HeadlessClient).Ctx.Err()
	}

	poolConfig := &pool.Config{
		InitialCap: cfg.InitialCap,
		MaxCap:     cfg.MaxCap,
		Factory:    factory,
		Close:      close,
		Ping:       ping,
		//连接最大空闲时间，超过该时间的连接 将会关闭，可避免空闲时连接EOF，自动失效的问题
		IdleTimeout: time.Duration(cfg.IdleTimeout) * time.Second,
	}

	return pool.NewChannelPool(poolConfig)
}
