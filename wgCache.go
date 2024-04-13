package wgCache

import (
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
	}
	groups[name] = g
	return g
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
	//调用getLocally在本地找该数据
	//如果是分布式场景，则需要调用getFromPeer函数从其他节点取（还没写）
	return g.getFromLocal(key)
}

func (g *Group) getFromLocal(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{arr: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

// 访问过后，将数据添加到缓存中
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
