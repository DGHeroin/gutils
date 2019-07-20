package gutils

import (
    "github.com/go-redis/redis"
    "time"
)

type RedisShareLimiter struct {
    client *redis.Client
    bucket int
}

func NewRedisShareLimiter(bucket int, addr string, password string, db int) (*RedisShareLimiter, error) {
    r := &RedisShareLimiter{}
    r.bucket = bucket
    r.client = redis.NewClient(&redis.Options{
        Addr:addr,
        Password:password,
        DB:db,
    })
    if err := r.client.Ping().Err(); err != nil {
        return nil, err
    }
    return r, nil
}

func (r *RedisShareLimiter) Allow(token string) (bool, error) {
    cmd := r.client.Decr(token)
    val, err := cmd.Result()
    if err == redis.Nil { // 还没有值, 重置初始值
        var incr *redis.IntCmd
        _, err := r.client.Pipelined(func(pipeliner redis.Pipeliner) error {
            incr = pipeliner.IncrBy(token, int64(r.bucket - 1))
            pipeliner.Expire(token, time.Second)
            return nil
        })
        rate, err := incr.Result()

        return rate > 0, err
    }
    if val > 0 {
        return true, err
    } else { // 令牌不够
        return false, nil
    }
}

