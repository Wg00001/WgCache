package enviction

import (
	"container/list"
	"errors"
	"log"
)

// 缓存删除方式
const (
	LRU = iota
	FIFO
	LFU
)

// Cache 缓存
type Cache struct {
	maxByte      int64                         //允许使用的最大内存
	nbytes       int64                         //当前内存
	linkedList   *list.List                    //双向链表用于实现回收算法，element.Value = entry
	view         map[string]*list.Element      //map用于快速访问，值为双向链表中对应节点的指针
	policy       uint8                         //回收策略
	remover      Remover                       //用来操作linkedlist以实现内存回收策略
	BeforeRemove func(key string, value Value) //回调函数，删除记录时执行，可以为空
}

// 双向链表节点的数据类型
type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}

// Remover 缓存删除算法的抽象
type Remover interface {
	Get(c *Cache, key *list.Element)
	Remove(c *Cache)
	Add(c *Cache, e *entry) *list.Element
}

func NewCache(maxByte int64, beforeRemove ...func(string, Value)) *Cache {
	if len(beforeRemove) > 1 {
		log.Println("Warning:view.struct.New:too many func")
	}
	return &Cache{
		maxByte:      maxByte,
		linkedList:   list.New(),
		view:         make(map[string]*list.Element),
		BeforeRemove: beforeRemove[0],
		policy:       LRU,
		remover:      new(lru),
	}
}

// SetPolicy 设置缓存删除方式
func (c *Cache) SetPolicy(tactics int) error {
	if tactics != FIFO && tactics != LFU && tactics != LRU {
		return errors.New("ERR:WgCache.view.struct.SetTactics: The policy does not exist")
	}
	c.policy = uint8(tactics)

	if c.policy == LRU {
		c.remover = new(lru)
	}

	return nil
}

// Add 新增或修改
func (c *Cache) Add(key string, value Value) {
	element, ok := c.view[key]
	if ok { //如果存在，则修改
		//相当于访问了一次
		c.remover.Get(c, element)
		kv := element.Value.(*entry)
		//修改时记得改当前内存长度
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		newEle := c.remover.Add(c, &entry{
			key:   key,
			value: value,
		})
		c.view[key] = newEle
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxByte != 0 && c.maxByte < c.nbytes {
		c.RemoveOldest()
	}
}

// Get 获取节点，同时通知remover
func (c *Cache) Get(key string) (v Value, ok bool) {
	if element, ok := c.view[key]; ok {
		kv := element.Value.(*entry)
		//通知remover
		c.remover.Get(c, element)
		return kv.value, true
	}
	return nil, false
}

func (c *Cache) RemoveOldest() {
	element := c.linkedList.Back()
	if element != nil {
		c.linkedList.Remove(element)
		kv := element.Value.(*entry)
		delete(c.view, kv.key)
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.BeforeRemove != nil {
			c.BeforeRemove(kv.key, kv.value)
		}
	}
}

func (c *Cache) Len() int {
	return c.linkedList.Len()
}
