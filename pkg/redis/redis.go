package redis

import (
	"crypto/tls"
	"sync"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"github.com/tsel-ticketmaster/tm-order/config"
)

var (
	client         redis.UniversalClient
	clientSyncOnce sync.Once
)

func buildConnection() redis.UniversalClient {
	cfg := config.Get()
	c := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:    cfg.Redis.Addrs,
		Username: cfg.Redis.Username,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	})

	redisotel.InstrumentTracing(c)
	redisotel.InstrumentMetrics(c)

	return c
}

func GetClient() redis.UniversalClient {
	clientSyncOnce.Do(func() {
		client = buildConnection()
	})

	return client
}
