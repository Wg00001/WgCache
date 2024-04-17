package wgCacheHttp

import (
	wgCache "WgCache"
	"WgCache/hash"
	protocol "WgCache/wgCacheProto"
	"fmt"
	"github.com/golang/protobuf/proto"
	"log"
	"net/http"
	"strings"
	"sync"
)

const (
	defaultPath     = "/wgCache/"
	defaultReplicas = 50
)

type HTTPPool struct {
	selfPath    string //自己的地址
	basePath    string //节点间通讯地址的前缀
	mu          sync.Mutex
	peers       *hash.Map
	httpGetters map[string]*httpGetter
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		selfPath: self,
		basePath: defaultPath,
	}
}

func (p *HTTPPool) Log(formate string, v ...interface{}) {
	log.Printf("[Server %s]: %s", p.selfPath, fmt.Sprintf(formate, v...))
}

func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("ERR:wgCacheHttp.ServeHTTP:HTTPPool serving unexpected path:" + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)
	//约定访问路径格式为：/<basepath>/<groupname>/<key>
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	groupName := parts[0]
	key := parts[1]

	group := wgCache.GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}
	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	body, err := proto.Marshal(&protocol.Response{Value: view.ByteSlice()})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(body)
}

// Set 更新节点列表
func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = hash.New(defaultReplicas)
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

func (p *HTTPPool) PickPeer(key string) (wgCache.DataGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	//调用一致性哈希路由表获取对应节点的HTTP客户端
	if peer := p.peers.Get(key); peer != "" && peer != p.selfPath {
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}
