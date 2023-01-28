package simpleCache

import "simpleCache/pb"

// 默认的实现在http.go
// 使用者可以自己实现对应的接口达到扩展功能的目的
// 例如使用其他通信协议等

// PeerPicker 本地服务端需要实现
// 在本地缓存未命中,且key不由本地负责时找到对该key负责的远端缓存
type PeerPicker interface {
	PickPeer(key string) (PeerGetter, bool)
}

// PeerGetter 本地客户端需要实现
// 从指定的peer中获取相应的数据
type PeerGetter interface {
	GetDataFromPeer(in *pb.Request, out *pb.Response) error
}
