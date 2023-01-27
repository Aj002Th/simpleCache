package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

/* 一致性哈希实现
 * 用于实现peers节点之间的负载均衡
 */

// HashFunc 定义哈希函数的原型
type HashFunc func([]byte) uint32

type Map struct {
	replicas int   // 用于控制生成的虚拟节点数目
	keys     []int // 排好序,模拟哈希环
	hashMap  map[int]string
	hashFunc HashFunc
}

func New(replicas int, hash HashFunc) *Map {
	if hash == nil {
		hash = crc32.ChecksumIEEE
	}
	return &Map{
		replicas: replicas,
		keys:     make([]int, 0),
		hashMap:  make(map[int]string),
		hashFunc: hash,
	}
}

// Add 向哈希环中插入peer的key
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hashKey := int(m.hashFunc([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hashKey)
			m.hashMap[hashKey] = key
		}
	}
	sort.Ints(m.keys)
}

// Get 通过想要得到的缓存内容key,得到未该key负责的peer
func (m *Map) Get(key string) string {
	hashKey := m.hashFunc([]byte(key))
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= int(hashKey)
	})

	return m.hashMap[m.keys[idx%len(m.keys)]]
}
