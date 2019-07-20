package gutils

import (
    "github.com/gin-gonic/gin"
    "golang.org/x/time/rate"
    "net/http"
    "sync"
    "time"
)

type ginVisitor struct {
    limiter *rate.Limiter
    lastSeen time.Time
}

type GinLimiter struct {
    requestsLimiter map[interface{}] *ginVisitor
    mutex sync.RWMutex
    limitPerSecond int
}

func NewGinLimiter(bucket int) *GinLimiter {
    g :=  &GinLimiter{}
    g.limitPerSecond = bucket //每1/r秒, 加入n个令牌
    g.requestsLimiter = make(map[interface{}]*ginVisitor)
    return g
}

//根据 IP 限制请求
func (limit *GinLimiter) Allow() gin.HandlerFunc {
    return func(c *gin.Context) {
        address := c.Request.RemoteAddr

        limit.mutex.RLock()
        r, ok := limit.requestsLimiter[address]
        limit.mutex.RUnlock()
        if !ok || r == nil {
            r = &ginVisitor{}
            r.limiter = rate.NewLimiter(1, limit.limitPerSecond)
            limit.mutex.Lock()
            limit.requestsLimiter[address] = r
            limit.mutex.Unlock()
        }
        if !r.limiter.Allow() {
            c.JSON(http.StatusTooManyRequests, gin.H{"code":-1, "err":http.StatusText(http.StatusTooManyRequests)})
            c.Abort()
            return
        }
        r.lastSeen = time.Now() // 上一次有效请求
        c.Next()
    }

}
