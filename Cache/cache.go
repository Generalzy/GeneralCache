package Cache

import (
	"github.com/Generalzy/GeneralCache/Cache/lru"
	"sync"
)

// cache 封装lru
type cache struct {
	// 互斥锁
	mu sync.RWMutex
	// lru 封装的lru缓存
	lruCache *lru.Cache
	// maxBytes 最大允许使用内存
	maxBytes int64
}

// add 封装了Add方法
func (c *cache) addKeyToCache(key string, value ReadOnlyByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// 懒加载lru.Cache
	if c.lruCache == nil {
		c.lruCache = lru.NewCache(c.maxBytes, nil)
	}
	c.lruCache.AddKeyToLruCache(key, value)
}

// get 封装了Get方法
func (c *cache) getKeyFromCache(key string) (ReadOnlyByteView, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	// 若未初始化就获取值则返回nil
	if c.lruCache == nil {
		return ReadOnlyByteView{}, false
	}

	if v, ok := c.lruCache.GetKeyFromLruCache(key); ok {
		// ByteView实现了Len接口
		// 因此v类型断言为ByteView
		return v.(ReadOnlyByteView), ok
	}

	return ReadOnlyByteView{}, false
}
