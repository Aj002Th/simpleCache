package simpleCache

// ByteView 作为存储在缓存中的一种 Value
// 特性是只读
type ByteView struct {
	b []byte // 可以支持如图片、视频等二进制数据
}

func (v ByteView) Len() int {
	return len(v.b)
}

// string天然不可修改,不用特殊处理
// 将数据看作string处理,便于缓存的字符串数据的使用
func (v ByteView) String() string {
	return string(v.b)
}

// ByteSlices 为了实现只读,需要进行深拷贝
func (v ByteView) ByteSlices() []byte {
	data := make([]byte, v.Len())
	copy(data, v.b)
	return data
}
