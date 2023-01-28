package simpleCache

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"simpleCache/consistenthash"
	"simpleCache/pb"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/_simplecache"
	defaultReplicas = 50
)

// HttpGetter http客户端
type HttpGetter struct {
	basePath string
}

func NewHttpGetter(basePath string) *HttpGetter {
	return &HttpGetter{basePath: basePath}
}

func (g *HttpGetter) GetDataFromPeer(in *pb.Request, out *pb.Response) error {
	peerUrl := fmt.Sprintf(
		"%s/%s/%s",
		g.basePath,
		url.QueryEscape(in.GetGroup()), // url转义保护
		url.QueryEscape(in.GetKey()),
	)
	resp, err := http.Get(peerUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("get data from %s failed with status %d", peerUrl, resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// 将得到的结果反序列化到out中
	if err = proto.Unmarshal(data, out); err != nil {
		return fmt.Errorf("decoding response body failed: %v", err)
	}

	return nil
}

// HttpPool http服务端
type HttpPool struct {
	// 服务端本地信息
	self     string
	basePath string

	// 用于请求远端缓存所需的信息
	mu          sync.Mutex
	peers       *consistenthash.Map
	httpGetters map[string]*HttpGetter
}

func NewHttpPool(self string) *HttpPool {
	return &HttpPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

// Set 会把整个peers设置更新,不保留原数据
func (p *HttpPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.peers = consistenthash.New(defaultReplicas, nil)
	p.peers.Add(peers...)

	// peer的值得是ip+端口
	p.httpGetters = make(map[string]*HttpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = NewHttpGetter(peer + p.basePath)
	}
}

func (p *HttpPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	peerKey := p.peers.Get(key)
	// 如果找不到负责的peer(这通常是出错了)或是自己负责
	// PickPeer调用失败,返回false要求本地执行回调去获取数据
	if peerKey == "" || peerKey == p.self {
		return nil, false
	}
	p.Log("Pick peer %s", peerKey)
	return p.httpGetters[peerKey], true
}

func (p *HttpPool) Log(format string, v ...any) {
	log.Printf("[server: %s] %s", p.self, fmt.Sprintf(format, v...))
}

// peer节点之间使用http协议进行通信
// 路径规则：ip:port/basePath/groupName/key
func (p *HttpPool) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	p.Log("%s - %s", req.Method, path)
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

	resp, err := proto.Marshal(&pb.Response{Value: data.ByteSlice()})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// 将数据当成普通二进制传输
	w.Header().Set("Content-Type", "application/octet-stream")
	_, _ = w.Write(resp)
}
