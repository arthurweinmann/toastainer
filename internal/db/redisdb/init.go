package redisdb

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/toastate/toastainer/internal/config"
	"github.com/toastate/toastainer/internal/nodes"
)

type Client interface {
	redis.Scripter
	redis.Cmdable
}

var client *redis.Client
var clusterClient *redis.ClusterClient

func GetClient() Client {
	if clusterClient != nil {
		return clusterClient
	}

	return client
}

func Init() (err error) {
	if config.NodeDiscovery {
		ips := nodes.GetAllLocalRedisNodes()
		if len(ips) == 0 {
			return fmt.Errorf("no local redis instance found")
		}

		if len(ips) == 1 {
			client = redis.NewClient(&redis.Options{
				Network:  "tcp",
				Addr:     ips[0], // port is declared in the ip directly
				Username: config.Redis.Username,
				Password: config.Redis.Password,
				DB:       config.Redis.DB,
			})
		} else {
			clusterClient = redis.NewClusterClient(&redis.ClusterOptions{
				// RouteByLatency: true,
				RouteRandomly: true,
				Addrs:         ips,
				NewClient: func(opt *redis.Options) *redis.Client {
					opt.Username = config.Redis.Username
					opt.Password = config.Redis.Password
					return redis.NewClient(opt)
				},
			})
		}
	} else {
		if len(config.Redis.IP) == 1 {
			client = redis.NewClient(&redis.Options{
				Network:  "tcp",
				Addr:     config.Redis.IP[0],
				Username: config.Redis.Username,
				Password: config.Redis.Password,
				DB:       config.Redis.DB,
			})
		} else {
			addrs := make([]string, len(config.Redis.IP))
			for i := 0; i < len(addrs); i++ {
				addrs[i] = config.Redis.IP[i]
			}
			clusterClient = redis.NewClusterClient(&redis.ClusterOptions{
				// RouteByLatency: true,
				RouteRandomly: true,
				Addrs:         addrs,
				NewClient: func(opt *redis.Options) *redis.Client {
					opt.Username = config.Redis.Username
					opt.Password = config.Redis.Password
					return redis.NewClient(opt)
				},
			})
		}
	}

	for i := 0; i < 3; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		err = GetClient().Ping(ctx).Err()
		cancel()
		if err == nil {
			return
		}
		time.Sleep(30 * time.Second)
	}

	return
}
