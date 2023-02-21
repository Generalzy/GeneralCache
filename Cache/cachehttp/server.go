package cachehttp

import (
	"fmt"
	"github.com/Generalzy/GeneralCache/Cache"
	"github.com/Generalzy/GeneralCache/Cache/consistenthash"
	"log"
	"net/http"
	"strings"
	"sync"
)

// 比较特殊的url前缀
// 举例: host:port/_general_cache/groupName/key 来获取某一个group的key
const (
	defaultReplicas = 50
	defaultBasePath = "/_general_cache/"
)

// HTTPServerPool node服务节点
type HTTPServerPool struct {
	// self 记录节点的ip和端口
	self string
	// baseUrl http的url前缀
	baseUrl string

	mu    sync.RWMutex
	nodes *consistenthash.Consistent
	// 存放http getter
	httpGetters map[string]*httpGetter
}

// NewHTTPServerPool initializes an HTTP pool of peers.
func NewHTTPServerPool(self string) *HTTPServerPool {
	return &HTTPServerPool{
		self:    self,
		baseUrl: defaultBasePath,
	}
}

// Log info with server name
func (p *HTTPServerPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s \n", p.self, fmt.Sprintf(format, v...))
}

func (p *HTTPServerPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 前缀不是group通信前缀返回400
	if !strings.HasPrefix(r.URL.Path, p.baseUrl) {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// 打印访问日志
	p.Log("%s %s", r.Method, r.URL.Path)

	// 访问url /<basepath>/<groupname>/<key> required
	parts := strings.SplitN(r.URL.Path[len(p.baseUrl):], "/", 2)

	// url不合规返回400
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// 根据切片获取group和key信息
	groupName, key := parts[0], parts[1]

	// 获取group
	cacheGroup := Cache.GetCacheGroup(groupName)
	if cacheGroup == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	// 获取val
	readonlyView, err := cacheGroup.GetKeyFromCacheGroup(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 响应view的拷贝
	_, _ = w.Write(readonlyView.CloneReadOnlyByteView())
}

// AddNode 方法实例化一致性哈希算法，并且添加了传入的节点。
func (p *HTTPServerPool) AddNode(node ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.nodes = consistenthash.NewConsistent(defaultReplicas, nil)
	// 加入节点
	p.nodes.AddNodeToConsistent(node...)
	// 初始化分布式节点
	p.httpGetters = make(map[string]*httpGetter, len(node))
	// 存储node
	for _, n := range node {
		// 建立node和httpGetter的映射
		p.httpGetters[n] = &httpGetter{baseURL: n + p.baseUrl}
	}
}

// GetNode 包装了一致性哈希算法的 Get() 方法，根据具体的 key，选择节点，返回节点对应的 HTTP 客户端。
func (p *HTTPServerPool) GetNode(key string) (Cache.NodeGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if node := p.nodes.GetNodeFromConsistent(key); node != "" && node != p.self {
		p.Log("Pick node %s", node)
		return p.httpGetters[node], true
	}
	return nil, false
}

var _ Cache.NodePicker = (*HTTPServerPool)(nil)
