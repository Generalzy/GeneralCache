package Cache

// ReadOnlyByteView 保存字节的不可变视图。
type ReadOnlyByteView struct {
	b []byte
}

// Len returns the view's length
func (v ReadOnlyByteView) Len() int {
	return len(v.b)
}

// clone 拷贝功能
func clone(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}

// CloneReadOnlyByteView 只读拷贝
func (v ReadOnlyByteView) CloneReadOnlyByteView() []byte {
	return clone(v.b)
}

// String 实现string接口
func (v ReadOnlyByteView) String() string {
	return string(v.b)
}
