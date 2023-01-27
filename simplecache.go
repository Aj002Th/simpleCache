package simpleCache

import (
	"errors"
	"log"
	"sync"
)

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// Getter 由用户传入的回调函数:如何从数据源拉取数据
type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// Group 核心结构
type Group struct {
	name      string     // 命名空间
	getter    Getter     // 回调函数
	mainCache cache      // 属于这个group的缓存
	peers     PeerPicker // 以此获取远端缓存
}

func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("getter is nil")
	}

	mu.Lock()
	defer mu.Unlock()

	g := &Group{
		name:   name,
		getter: getter,
		mainCache: cache{
			cacheBytes: cacheBytes,
		},
	}

	groups[name] = g
	return g
}

// RegisterPeerPicker 相当于把PeerPicker的初始化从NewGroup中单独拉出来的
// 主要是考虑到这个PeerPicker可能比较复杂, 而且NewGroup参数列表已经很长了
// 这个函数每个Group只能调用一次
func (g *Group) RegisterPeerPicker(picker PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker have been called before")
	}
	g.peers = picker
}

// GetGroup 对应group不存在时返回nil
func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()

	g, ok := groups[name]
	if !ok {
		return nil
	}
	return g
}

// Get simpleCache对外服务的唯一接口
// 若本地缓存命中,则从本地缓存中获取数据 -> getLocally
// 本地缓存未命中,且对应key不由本地缓存负责时,请求对应的远程缓存来获取数据 -> getRemote
// 本地缓存未命中,且对应key由本地缓存负责时,调用用户传入的回调函数从数据源获取数据 -> g.getter.Get
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, errors.New("get a empty key")
	}

	data, ok := g.mainCache.get(key)
	if ok {
		return data, nil
	}
	return g.load(key)
}

// 缓存未命中时的处理
func (g *Group) load(key string) (ByteView, error) {
	// 先只实现单机版, 不会从远端缓存拉取数据
	peerGetter, ok := g.peers.PickPeer(key)
	if ok {
		// 请求远端缓存获取数据
		data, err := g.getFromPeer(peerGetter, key)
		if err == nil {
			return data, nil
		}
		// 失败了就打日志+本地执行回调
		log.Printf("get data(key:%s) from peer failed: %v", key, err)
	}

	// 本地调用回调获取数据
	return g.getLocally(key)
}

// 本地调用回调函数从数据源获取数据
func (g *Group) getLocally(key string) (ByteView, error) {
	data, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}

	value := ByteView{b: data}
	g.populateCache(key, value)
	return value, nil
}

// 从peer获取数据
func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	data, err := peer.GetDataFromPeer(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: data}, nil
}

// 向group的缓存中添加数据
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
