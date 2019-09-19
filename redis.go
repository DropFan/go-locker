package locker

import (
	"context"
	"errors"

	"github.com/go-redis/redis"
)

var (
	// rc redis client
	rc redis.Cmdable
)

// SetRedis set redis client (go-redis)
func SetRedis(ctx context.Context, r redis.Cmdable) {
	rc = r
}

// InitRedis init redis client (go-redis)
func InitRedis(ctx context.Context, opts *redis.ClusterOptions) {
	rc = redis.NewClusterClient(opts)
}

// GoRedisEvalFunc ...eval func warpper for go-redis
func GoRedisEvalFunc(ctx context.Context, script string, keys []string, args ...interface{}) (ret int, err error) {
	if rc == nil {
		return -1, errors.New("redis client is nil")
	}
	// var args []interface{}

	// for val := range values {
	// 	args = append(args, val)
	// }

	res := rc.Eval(script, keys, args...)

	ret, err = res.Int()

	if err != nil {
		return -1, err
	}

	return
}
