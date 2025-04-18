package hw04lrucache

type Key string

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
}

type lruCache struct {
	Cache // Remove me after realization.

	capacity int
	queue    List
	items    map[Key]*ListItem
}

type cacheListItem struct {
	key   Key
	value interface{}
}

func NewCache(capacity int) Cache {
	if capacity < 1 {
		return nil
	}

	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
	}
}

func (c *lruCache) Set(key Key, value interface{}) bool {
	listItem := &cacheListItem{key, value}

	// The element is present in the cache -> updating it's value, moving it to the front.
	if v, ok := c.items[key]; ok {
		v.Value = listItem
		c.queue.MoveToFront(v)
		return true
	}

	newElem := c.queue.PushFront(listItem)
	c.items[key] = newElem

	if c.queue.Len() > c.capacity {
		delete(c.items, c.queue.Back().Value.(*cacheListItem).key)
		c.queue.Remove(c.queue.Back())
	}

	return false
}

func (c *lruCache) Get(key Key) (interface{}, bool) {
	if v, ok := c.items[key]; ok {
		c.queue.MoveToFront(v)
		return v.Value.(*cacheListItem).value, true
	}

	return nil, false
}
