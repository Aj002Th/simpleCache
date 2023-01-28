package singleflight

import "sync"

// Group 将同样的缓存请求合并成一个
type Group struct {
	mu      sync.Mutex
	callers map[string]*caller
}

type caller struct {
	wg  sync.WaitGroup
	val any
	err error
}

func (g *Group) Do(id string, fn func() (any, error)) (any, error) {
	g.mu.Lock()

	// 延迟初始化
	if g.callers == nil {
		g.callers = make(map[string]*caller)
	}

	c, ok := g.callers[id]
	if ok {
		g.mu.Unlock() // 这里解锁可以让后续的重复查询也进入到下一行阻塞等结果
		c.wg.Wait()
		return c.val, c.err
	}

	// 首个查询才会真正执行
	c = new(caller)
	c.wg.Add(1)
	g.callers[id] = c
	g.mu.Unlock()

	c.val, c.err = fn()
	c.wg.Done()

	// 删除标记
	g.mu.Lock()
	delete(g.callers, id)
	g.mu.Unlock()

	return c.val, c.err
}
