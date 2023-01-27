package simpleCache

import (
	"simpleCache/lru"
	"sync"
)

// 其实就是对lru中的cache再包装了一层,增加了并发访问控制
// 并且将cache中的value指定为了byteView
type cache struct {
	mu         sync.Mutex // 实现并发控制
	lru        *lru.Cache // 实际存储信息的位置
	cacheBytes int64      // 控制缓存空间的大小, 0时不进行限制
}

func (c *cache) get(key string) (ByteView, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lru == nil {
		// todo:这里onEvict没有给设置的方法
		c.lru = lru.New(c.cacheBytes, nil)
	}

	val, ok := c.lru.Get(key)
	if !ok {
		return ByteView{}, false
	}
	return val.(ByteView), true
}

func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lru == nil {
		// todo:这里onEvict没有给设置的方法
		c.lru = lru.New(c.cacheBytes, nil)
	}

	c.lru.Add(key, value)
}
