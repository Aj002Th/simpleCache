package lru

import "container/list"

type Cache struct {
	// 用于管理缓存空间的大小
	// maxbytes设置为0则不对缓存空间大小进行限制
	maxbytes int64
	nbytes   int64

	// 用选择淘汰数据
	// 从队尾开始淘汰
	ll *list.List

	// 快速查找缓存数据
	cache map[string]*list.Element

	// 一个钩子函数,可以自行设置数据被淘汰时还有什么额外工作需要做
	OnEvict func(key string, val Value)
}

func New(maxbytes int64, onEvict func(key string, val Value)) *Cache {
	return &Cache{
		maxbytes: maxbytes,
		nbytes:   0,
		ll:       list.New(),
		cache:    make(map[string]*list.Element),
		OnEvict:  onEvict,
	}
}

// 链表节点中存储的数据
type entry struct {
	key string
	val Value
}

// Value 用于计算缓存数据的大小
type Value interface {
	Len() int
}

// Get 从cache中读数据
func (c *Cache) Get(key string) (Value, bool) {
	if ele, ok := c.cache[key]; ok {
		// 移到队头
		c.ll.MoveToFront(ele)

		kv := ele.Value.(*entry)
		return kv.val, true
	}

	// 未命中
	return nil, false
}

// RemoveOldest 内存淘汰
func (c *Cache) RemoveOldest() {
	outEle := c.ll.Back()
	if outEle == nil {
		return
	}

	c.ll.Remove(outEle)
	kv := outEle.Value.(*entry)
	c.nbytes -= int64(len(kv.key)) + int64(kv.val.Len())
	delete(c.cache, kv.key)

	if c.OnEvict != nil {
		c.OnEvict(kv.key, kv.val)
	}
}

// Add 放入缓存
func (c *Cache) Add(key string, value Value) {
	// 避免一个特大的数据把缓存中的数据清空了
	// 拒绝缓存这样的数据
	totalBytes := int64(len(key)) + int64(value.Len())
	if totalBytes > c.maxbytes && c.maxbytes != 0 {
		return
	}

	if oldKV, ok := c.cache[key]; ok {
		// 找到就修改
		c.ll.MoveToFront(oldKV)
		ele := oldKV.Value.(*entry)
		ele.val = value
	} else {
		// 没找到就新建
		kv := &entry{
			key: key,
			val: value,
		}
		ele := c.ll.PushFront(kv)
		c.nbytes += totalBytes
		c.cache[key] = ele
	}

	// 内存淘汰
	for c.maxbytes != 0 && c.nbytes > c.maxbytes {
		c.RemoveOldest()
	}
}

// Len 缓存条数
func (c *Cache) Len() int {
	return c.ll.Len()
}
