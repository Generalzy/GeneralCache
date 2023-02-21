package Cache

// NodePicker 的 GetNode() 方法用于根据传入的 key 选择相应节点 NodeGetter。
type NodePicker interface {
	GetNode(key string) (NodeGetter, bool)
}

// NodeGetter 的 Get() 方法用于从对应 group 查找缓存值。NodeGetter 就对应于上述流程中的 HTTP 客户端。
type NodeGetter interface {
	GetKeyFromGetter(group string, key string) ([]byte, error)
}
