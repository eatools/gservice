package gcache

import (
	"log"
	"os"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/mna/redisc"
)

func InitGcache(address []string, passwd string) (g *GCache) {
	if len(address) == 1 {
		g = &GCache{poolClient: nPool(address[0], passwd)}
	} else if len(address) > 1 {
		g = &GCache{clusterClient: newCluster(address, passwd)}
	}

	return
}

func nPool(addr, passwd string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     50,  // 最大空闲连结数
		MaxActive:   256, // 最大连接数
		IdleTimeout: 3 * time.Second,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", addr)
			// Connection error handling
			if err != nil {
				log.Printf("ERROR: fail initializing the redis pool: %s", err.Error())
				os.Exit(1)
			}
			if passwd != "" {
				if _, err := conn.Do("AUTH", passwd); err != nil {
					log.Printf("ERROR: fail password the redis pool: %s", err.Error())
					conn.Close()
					os.Exit(1)
				}
			}
			return conn, err
		},
	}
}

func newCluster(address []string, passwd string) *redisc.Cluster {
	// create the cluster
	cluster := &redisc.Cluster{
		StartupNodes: address,
		DialOptions:  []redis.DialOption{redis.DialConnectTimeout(5 * time.Second)},
		CreatePool: func(addr string, options ...redis.DialOption) (*redis.Pool, error) {
			return nPool(addr, passwd), nil
		},
	}
	return cluster
}

type GCache struct {
	poolClient    *redis.Pool
	clusterClient *redisc.Cluster
}

func (g *GCache) Get() redis.Conn {
	if g.clusterClient != nil {
		return g.clusterClient.Get()
	}
	if g.poolClient != nil {
		return g.poolClient.Get()
	}
	return nil
}

func (g *GCache) Find(key string) (uint64, error) {
	redisConn := g.Get()
	defer redisConn.Close()

	return redis.Uint64(redisConn.Do("get", key))
}
