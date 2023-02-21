package lru

import (
	"container/list"
	"fmt"
)

// Callback 回调函数
type Callback func(key string, value Value)

// Cache LRU缓存
type Cache struct {
	// maxBytes 最大允许使用内存
	maxBytes int64
	// currentBytes 当前使用内存
	currentBytes int64

	// linker 底层链表
	linker *list.List
	// cache 底层缓存
	cache map[string]*list.Element

	// callback 某个key被移除后的回调函数
	callback Callback
}

func (c *Cache) String() string {
	return fmt.Sprintf("cacheInfo: currentBytes: %d | maxBytes: %d", c.currentBytes, c.maxBytes)
}

// Value 返回值所占用的内存大小
type Value interface {
	Len() int
}

// entry linker的node
type entry struct {
	key   string
	value Value
}

func NewCache(maxBytes int64, callback Callback) *Cache {
	return &Cache{
		maxBytes: maxBytes,
		linker:   list.New(),
		cache:    make(map[string]*list.Element),
		callback: callback,
	}
}

// GetKeyFromLruCache 查询key
func (c *Cache) GetKeyFromLruCache(key string) (Value, bool) {
	if val, ok := c.cache[key]; ok {
		// 移到队尾部
		c.linker.MoveToBack(val)
		// 将list.Element.Value类型断言为entry
		kv := val.Value.(*entry)
		return kv.value, true
	}
	return nil, false
}

// RemoveOldestKeyFromLru 内存淘汰机制
func (c *Cache) RemoveOldestKeyFromLru() {
	// 返回队首
	ele := c.linker.Front()
	if ele != nil {
		// 从链表中删除元素
		c.linker.Remove(ele)
		kv := ele.Value.(*entry)

		// 从cache中将key淘汰
		delete(c.cache, kv.key)

		// 修改当前cache占用大小
		// 即减去一个k,一个v的大小
		c.currentBytes -= int64(len(kv.key)) + int64(kv.value.Len())

		// 如果用户定义的回调函数不为空则执行一下
		if c.callback != nil {
			c.callback(kv.key, kv.value)
		}
	}
}

// AddKeyToLruCache adds a value to the cachehttp.
func (c *Cache) AddKeyToLruCache(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		// 修改节点

		// 移动到队尾
		c.linker.MoveToBack(ele)
		// 获取entry(key,val)
		kv := ele.Value.(*entry)
		// 当前内存占用为旧val长度-新val长度
		c.currentBytes += int64(value.Len()) - int64(kv.value.Len())
		// 覆盖旧value
		kv.value = value
	} else {
		// 从队尾加入
		ele := c.linker.PushBack(&entry{key, value})
		c.cache[key] = ele
		// 增加一个key和一个val的长度
		c.currentBytes += int64(len(key)) + int64(value.Len())
	}

	// 如果超过限制,则进行内存淘汰
	for c.maxBytes != 0 && c.maxBytes < c.currentBytes {
		c.RemoveOldestKeyFromLru()
	}
}

func (c *Cache) Len() int {
	return c.linker.Len()
}
