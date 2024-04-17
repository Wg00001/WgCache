package wgCache

import (
	"WgCache/singleFlight"
	protocol "WgCache/wgCacheProto"
	"fmt"
	"log"
	"sync"
)

//负责与外部交互，控制缓存存储和获取的主流程

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// Group 相当于一个缓存的命名空间
type Group struct {
	name      string //每个Group一个唯一的name
	getter    Getter //缓存未命中时获取源数据的回调函数
	mainCache cache  //缓存
	peers     PeerPicker
	loader    *singleFlight.Group //防止缓存击穿
}

// NewGroup 创建新group
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("ERR:wgCache.wgCache.NewGroup:nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
		loader:    &singleFlight.Group{},
	}
	groups[name] = g
	return g
}

// RegisterPeers 用于将PeerPicker接口的HTTPPool注入到Group中
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker = nil")
	}
	g.peers = peers
}

// GetGroup 返回之前使用NewGroup创建的命名组，如果没有这样的组，则返回nil。
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}
	if v, ok := g.mainCache.get(key); ok {
		log.Println("wgCache get ", key)
		return v, nil
	}
	return g.load(key)
}

func (g *Group) load(key string) (value ByteView, err error) {
	//如果是分布式场景，则需要调用getFromPeer函数从其他节点取
	viewi, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[wgCache] Failed to get from peer", err)
			}
		}
		//调用getLocally在本地找该数据
		return g.getFromLocal(key)
	})
	if err == nil {
		return viewi.(ByteView), nil
	}
	return
}

func (g *Group) getFromLocal(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{bytes: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

// 从使用http访问远程节点以获取缓存值
func (g *Group) getFromPeer(peer DataGetter, key string) (ByteView, error) {
	req := &protocol.Request{
		Group: g.name,
		Key:   key,
	}
	rsp := &protocol.Response{}
	err := peer.Get(req, rsp)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{bytes: rsp.Value}, nil
}

// 访问过后，将数据添加到缓存中
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
