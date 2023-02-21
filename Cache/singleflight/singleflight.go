package singleflight

import "sync"

// request 一次请求
type request struct {
	wg  sync.WaitGroup
	val any
	err error
}

// RequestGroup 管理不同 key 的请求request
type RequestGroup struct {
	mu sync.RWMutex
	m  map[string]*request
}

func (g *RequestGroup) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	// 加锁:map不是线程安全的
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*request)
	}

	if req, ok := g.m[key]; ok {
		// 如果request存在,则等待执行完成
		g.mu.Unlock()
		req.wg.Wait()
		return req.val, req.err
	}

	// new一个request
	// 此处为指针变量,便于后续修改request的值
	req := new(request)
	// wg计数器加一
	req.wg.Add(1)
	// 将当前key的request存入group
	g.m[key] = req
	// 操作结束解锁
	g.mu.Unlock()

	// 调用fn获取结果
	req.val, req.err = fn()
	// 计数器减一
	req.wg.Done()

	// 加锁处理map
	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return req.val, req.err
}
