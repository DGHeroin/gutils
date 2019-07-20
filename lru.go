package gutils


import (
    "container/list"
    "sync"
)

const (
    RemoveTypeFullEntries = RemoveReason(0)
    RemoveTypeFullMemory  = RemoveReason(1)
    RemoveTypeByUser      = RemoveReason(2)
)

type Cache struct {
    MaxEntries  uint64
    MaxMemory   uint64
    memoryCount uint64
    OnEvicted   func(key string, value []byte, fullType RemoveReason)
    ll          *list.List
    cache       map[interface{}]*list.Element
    mu          sync.RWMutex
}

type RemoveReason int

type entry struct {
    key   string
    value []byte
}

func New(maxEntries uint64, maxMemory uint64) *Cache {
    if maxEntries == 0 {
        maxEntries= ^uint64(0)
    }
    if maxMemory == 0 {
        maxMemory = ^uint64(0)
    }

    return &Cache{
        MaxEntries: maxEntries,
        MaxMemory:  maxMemory,
        ll:         list.New(),
        cache:      make(map[interface{}]*list.Element),
    }
}

func (c *Cache) Add(key string, value []byte) {
    c.mu.Lock()
    defer c.mu.Unlock()
    if c.cache == nil {
        c.cache = make(map[interface{}]*list.Element)
        c.ll = list.New()
    }
    itemSize := uint64(len(value))
    // check memory enough
    if itemSize > c.MaxMemory {
        // too large
        return
    }

    for {
        left := c.MaxMemory - c.memoryCount
        if c.lenLocked() == 0 {
            // remove all but memory still not enough
            if left > itemSize {
                break
            }
            return
        }
        if left < itemSize {
            c.removeOldestLocked(RemoveTypeFullMemory)
            continue
        }
        break
    }

    if ee, ok := c.cache[key]; ok {
        c.ll.MoveToFront(ee)
        ee.Value.(*entry).value = value
        c.memoryCount += itemSize
        return
    }
    c.memoryCount += itemSize
    ele := c.ll.PushFront(&entry{key, value})
    c.cache[key] = ele
    if uint64(c.ll.Len()) > c.MaxEntries {
        c.removeOldestLocked(RemoveTypeFullEntries)
    }
}

func (c *Cache) Get(key string) (value []byte, ok bool) {
    c.mu.Lock()
    defer c.mu.Unlock()
    if c.cache == nil {
        return
    }
    if ele, hit := c.cache[key]; hit {
        c.ll.MoveToFront(ele)
        return ele.Value.(*entry).value, true
    }
    return
}

func (c *Cache) Remove(key string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    if c.cache == nil {
        return
    }
    if ele, hit := c.cache[key]; hit {
        c.removeElement(ele, RemoveTypeByUser)
    }
}

func (c *Cache) RemoveOldest() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.removeOldestLocked(RemoveTypeByUser)
}

func (c *Cache) removeOldestLocked(removeType RemoveReason) {
    if c.cache == nil {
        return
    }
    ele := c.ll.Back()
    if ele != nil {
        if data, ok := ele.Value.(*entry); ok {
            c.memoryCount -= uint64(len(data.value))
        }
        c.removeElement(ele, removeType)
    }
}

func (c *Cache) removeElement(e *list.Element, removeType RemoveReason) {
    c.ll.Remove(e)
    kv := e.Value.(*entry)
    delete(c.cache, kv.key)
    if c.OnEvicted != nil {
        c.OnEvicted(kv.key, kv.value, removeType)
    }
}

func (c *Cache) Len() uint64 {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.lenLocked()
}

func (c *Cache) lenLocked() uint64 {
    if c.cache == nil {
        return 0
    }
    return uint64(c.ll.Len())
}


func (c *Cache) Clear() {
    c.mu.Lock()
    defer c.mu.Unlock()
    if c.OnEvicted != nil {
        for _, e := range c.cache {
            kv := e.Value.(*entry)
            c.OnEvicted(kv.key, kv.value, RemoveTypeFullEntries)
        }
    }
    c.ll = nil
    c.cache = nil
}

func (r RemoveReason) String() string {
    switch r {
    case RemoveTypeFullEntries: return "Remove by full entries"
    case RemoveTypeFullMemory: return "Remove by full memory"
    case RemoveTypeByUser:return "Remove by user"
    }
    return "Unknown remove reason"
}
