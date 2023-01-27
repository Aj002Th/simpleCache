package simpleCache

import (
	"log"
	"net/http"
	"strings"
)

const defaultBasePath = "/_simplecache"

type HttpPool struct {
	self     string
	basePath string
}

func NewHttpPool(self string) *HttpPool {
	return &HttpPool{self, defaultBasePath}
}

func (p *HttpPool) Log(method, path string) {
	log.Printf("[server: %s] %s - %s", p.self, method, path)
}

// peer节点之间使用http协议进行通信
// 路径规则：ip:port/basePath/groupName/key
func (p *HttpPool) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	p.Log(req.Method, path)
	if !strings.HasPrefix(path, p.basePath) {
		http.Error(w, "do not match HttpPool's base path", 400)
		return
	}

	parts := strings.SplitN(path[len(p.basePath)+1:], "/", 2)
	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group:"+groupName, 400)
		return
	}

	data, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// 将数据当成普通二进制传输
	w.Header().Set("Content-Type", "application/octet-stream")
	_, _ = w.Write(data.ByteSlices())
}
