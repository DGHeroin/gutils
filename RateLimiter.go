package gutils

import (
    "golang.org/x/time/rate"
    "sync"
    "time"
)

type visitor struct {
    limiter  *rate.Limiter
    lastSeen time.Time
}

type RateLimiter struct {
    mutex    sync.RWMutex
    visitors map[interface{}]*visitor
    bucket   int
}

func NewRateLimiter(bucket int) *RateLimiter {
    r := &RateLimiter{}
    r.bucket = bucket
    r.visitors = make(map[interface{}]*visitor)
    return r
}

func (r *RateLimiter) Allow(i interface{}) bool {
    if i == nil { return false }
    r.mutex.RLock()
    v, ok := r.visitors[i]
    r.mutex.RUnlock()
    if !ok {
        v = &visitor{}
        v.limiter = rate.NewLimiter(1, r.bucket)
        r.mutex.Lock()
        r.visitors[i] = v
        r.mutex.Unlock()
    }
    allow := v.limiter.Allow()
    if !allow { return allow }
    v.lastSeen = time.Now()
    return allow
}
