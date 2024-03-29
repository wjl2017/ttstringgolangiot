package registry

import (
	"git.oschina.net/cloudzone/smartgo/stgnet/protocol"
)

// KVConfigSerializeWrapper KV配置的json序列化结构
// Author: tianyuliang
// Since: 2017/9/4
type KVConfigSerializeWrapper struct {
	ConfigTable map[string]map[string]string `json:"configTable"` // 数据格式：Namespace[Key[Value]]
	*protocol.RemotingSerializable
}

// NewKVConfigSerializeWrapper 初始化KV配置
// Author: tianyuliang
// Since: 2017/9/4
func NewKVConfigSerializeWrapper(configTable map[string]map[string]string) *KVConfigSerializeWrapper {
	kvConfigWrapper := &KVConfigSerializeWrapper{
		ConfigTable: configTable,
	}
	return kvConfigWrapper
}
