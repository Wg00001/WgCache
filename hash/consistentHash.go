package hash

import (
	"hash/crc32"
	"log"
	"sort"
	"strconv"
)

/*
Chord算法实现一致性哈希
*/

type Hash func(data []byte) uint32

// Map 路由表
type Map struct {
	hash     Hash //依赖注入，允许用于替换成自定义的Hash函数，默认为crc32.ChecksumIEEE算法
	replicas int  //虚拟节点倍数，虚拟节点用于扩充节点数量以解决数据倾斜问题
	keys     []int
	hashMap  map[int]string
}

// New 虚拟节点倍数需要自定义
func New(replicas int) *Map {
	return &Map{
		replicas: replicas,
		hash:     crc32.ChecksumIEEE,
		hashMap:  make(map[int]string),
	}
}

func (m *Map) SetHashFunc(hash Hash) {
	if hash == nil {
		log.Printf("hash func can't not be nil")
		return
	}
	m.hash = hash
}

// Add 增加节点，每个真实节点key创建replicas个虚拟节点
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

// Get 选择节点
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}
	hash := int(m.hash([]byte(key)))

	//使用二分查询搜索合适的节点
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})
	return m.hashMap[m.keys[idx%len(m.keys)]]
}
