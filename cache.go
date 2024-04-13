package wgCache

import (
	"WgCache/enviction"
	"sync"
)

//包装上互斥锁，用于支持并发读写。

type cache struct {
	mu         sync.Mutex
	cache      *enviction.Cache
	cacheBytes int64
}

func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	//延迟初始化：发现对象为空再初始化，不发现时无所谓
	if c.cache == nil {
		c.cache = enviction.NewCache(c.cacheBytes, nil)
	}
	c.cache.Add(key, value)
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cache == nil {
		return
	}
	if v, ok := c.cache.Get(key); ok {
		return v.(ByteView), ok
	}
	return
}
