package wgCache

//缓存值的抽象与封装

// ByteView 一个只读的数据结构，用来表示缓存值
type ByteView struct {
	arr []byte
}

func (v ByteView) Len() int {
	return len(v.arr)
}

func (v ByteView) String() string {
	return string(v.arr)
}

func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.arr)
}

func cloneBytes(b []byte) []byte {
	clone := make([]byte, len(b))
	copy(clone, b)
	return clone
}
