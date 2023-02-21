package Cache

import (
	"fmt"
	"github.com/Generalzy/GeneralCache/Cache/singleflight"
	"log"
	"sync"
)

// CacheGroup 对cache封装
type CacheGroup struct {
	// 当前组的名称
	groupName string
	// cacheGetter 外部加载key接口
	cacheGetter Getter
	// baseCache 底层缓存
	baseCache cache

	picker NodePicker
	// 请求组
	requestGroup *singleflight.RequestGroup
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*CacheGroup)
)

// NewCacheGroup 创建一个CacheGroup
func NewCacheGroup(groupName string, maxBytes int64, getter Getter) *CacheGroup {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &CacheGroup{
		groupName:   groupName,
		cacheGetter: getter,
		// 使用封装后的cache
		baseCache:    cache{maxBytes: maxBytes},
		requestGroup: new(singleflight.RequestGroup),
	}
	groups[groupName] = g
	return g
}

// GetCacheGroup 返回一个CacheGroup
func GetCacheGroup(groupName string) *CacheGroup {
	mu.RLock()
	defer mu.RUnlock()
	g := groups[groupName]
	return g
}

// GetKeyFromCacheGroup 从CacheGroup中获取key
func (g *CacheGroup) GetKeyFromCacheGroup(key string) (ReadOnlyByteView, error) {
	if key == "" {
		return ReadOnlyByteView{}, fmt.Errorf("[ERROR] key is required")
	}

	if v, ok := g.baseCache.getKeyFromCache(key); ok {
		log.Println("[INFO] hit local cache")
		return v, nil
	}

	// 从getter中获取数据
	return g.loadKeyFromGetter(key)
}

func (g *CacheGroup) loadKeyFromGetter(key string) (ReadOnlyByteView, error) {

	view, err := g.requestGroup.Do(key, func() (interface{}, error) {
		if g.picker != nil {
			if node, ok := g.picker.GetNode(key); ok {
				if value, err := g.getKeyFromNode(node, key); err == nil {
					return value, err
				}
			}
		}

		return g.getKeyFromLocal(key)
	})

	if err != nil {
		return ReadOnlyByteView{}, err
	}

	return view.(ReadOnlyByteView), nil
}

// RegisterPickerToCacheGroup registers a NodePicker for choosing remote peer
func (g *CacheGroup) RegisterPickerToCacheGroup(picker NodePicker) {
	if g.picker != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.picker = picker
}

func (g *CacheGroup) getKeyFromLocal(key string) (ReadOnlyByteView, error) {
	// 从getter中获取数据
	bytes, err := g.cacheGetter.Get(key)
	log.Printf(`[LOCAL INFO] get "%s" from "%s" local getter %s`, key, g.groupName, "\n")

	if err != nil {
		return ReadOnlyByteView{}, err
	}
	// 返回获取到数据的copy
	value := ReadOnlyByteView{b: clone(bytes)}
	// 加入缓存
	g.baseCache.addKeyToCache(key, value)
	return value, nil
}

func (g *CacheGroup) getKeyFromNode(getter NodeGetter, key string) (ReadOnlyByteView, error) {
	bytes, err := getter.GetKeyFromGetter(g.groupName, key)
	if err != nil {
		return ReadOnlyByteView{}, err
	}
	return ReadOnlyByteView{b: bytes}, nil
}
