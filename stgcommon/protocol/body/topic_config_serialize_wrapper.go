package body

import (
	"fmt"
	"git.oschina.net/cloudzone/smartgo/stgcommon"
	"git.oschina.net/cloudzone/smartgo/stgnet/protocol"
)

// TopicConfigSerializeWrapper topic
// Author gaoyanlei
// Since 2017/8/11
type TopicConfigSerializeWrapper struct {
	TopicConfigTable *TopicConfigTable      `json:"topicConfigTable"`
	DataVersion      *stgcommon.DataVersion `json:"dataVersion"`
	*protocol.RemotingSerializable
}

// NewTopicConfigSerializeWrapper 格式化
// Author: tianyuliang
// Since: 2017/10/21
func NewTopicConfigSerializeWrapper(dataVersion ...*stgcommon.DataVersion) *TopicConfigSerializeWrapper {
	topicConfigSerializeWrapper := &TopicConfigSerializeWrapper{
		TopicConfigTable:     NewTopicConfigTable(),
		RemotingSerializable: new(protocol.RemotingSerializable),
	}

	topicConfigSerializeWrapper.DataVersion = stgcommon.NewDataVersion()
	if dataVersion != nil && len(dataVersion) > 0 {
		topicConfigSerializeWrapper.DataVersion = dataVersion[0]
	}
	return topicConfigSerializeWrapper
}

// ToString 格式化
// Author: tianyuliang
// Since: 2017/10/21
func (self *TopicConfigSerializeWrapper) ToString() string {
	format := "TopicConfigSerializeWrapper [%s, %s]"
	return fmt.Sprintf(format, self.DataVersion.ToString(), self.TopicConfigTable.ToString())
}
