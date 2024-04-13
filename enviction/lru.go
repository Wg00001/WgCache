package enviction

import "container/list"

type lru struct {
}

var _ Remover = (*lru)(nil)

// 访问时，将节点移动到队头
func (l lru) Get(c *Cache, element *list.Element) {
	c.linkedList.MoveToFront(element)
}

func (l lru) Remove(cache *Cache) {
	//TODO implement me
	panic("implement me")
}

func (l lru) Add(c *Cache, e *entry) *list.Element {
	return c.linkedList.PushFront(e)
}
