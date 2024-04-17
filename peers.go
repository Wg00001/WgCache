package wgCache

import protocol "WgCache/wgCacheProto"

// PeerPicker 节点选取器，用于定位节点
type PeerPicker interface {
	// PickPeer 根据传入的key选择对应节点的PeerGetter
	PickPeer(key string) (DataGetter, bool)
}

// DataGetter 由节点实现
type DataGetter interface {
	// Get 从对应group中获取对应缓存值
	Get(in *protocol.Request, out *protocol.Response) error
}
