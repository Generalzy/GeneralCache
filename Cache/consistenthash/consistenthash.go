package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash 哈希函数
type Hash func(data []byte) uint32

// Consistent 一致性哈希
type Consistent struct {
	// hash 哈希函数
	hash Hash
	// replicas 节点倍数
	nodeReplicas int
	// ring 节点环
	nodeRing []int
	// virtualNodeMapping
	virtualNodeMapping map[int]string
}

// NewConsistent creates a Map instance
func NewConsistent(replicas int, fn Hash) *Consistent {
	m := &Consistent{
		nodeReplicas:       replicas,
		hash:               fn,
		virtualNodeMapping: make(map[int]string),
	}

	// 默认哈希函数
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// AddNodeToConsistent 向哈希环中添加节点
func (m *Consistent) AddNodeToConsistent(node ...string) {
	for _, key := range node {
		// 根据虚拟节点倍数添加虚拟节点
		// key: host1:6379 host2:6379 host3:6379
		for i := 0; i < m.nodeReplicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			// 加入hash环
			m.nodeRing = append(m.nodeRing, hash)
			// 添加虚拟节点和真实节点的映射
			m.virtualNodeMapping[hash] = key
		}
	}
	// 排序
	sort.Ints(m.nodeRing)
}

// GetNodeFromConsistent 获取节点
func (m *Consistent) GetNodeFromConsistent(node string) string {
	if len(m.nodeRing) == 0 {
		return ""
	}

	nodeHash := int(m.hash([]byte(node)))
	// Binary search for appropriate replica.
	idx := sort.Search(len(m.nodeRing), func(i int) bool {
		return m.nodeRing[i] >= nodeHash
	})

	return m.virtualNodeMapping[m.nodeRing[idx%len(m.nodeRing)]]
}
