package process

import (
	"git.oschina.net/cloudzone/smartgo/stgclient"
	"git.oschina.net/cloudzone/smartgo/stgcommon/logger"
	syncMap "git.oschina.net/cloudzone/smartgo/stgcommon/sync"
	"sync"
	"sync/atomic"
)

var (
	once     sync.Once
	instance *MQClientManager
)

// MQClientManager: MQClientInstance管理类
// Author: yintongqiang
// Since:  2017/8/8

type MQClientManager struct {
	// clientId MQClientInstance
	FactoryTable          *syncMap.Map
	FactoryIndexGenerator int32
}

func NewMQClientManager() *MQClientManager {
	return &MQClientManager{FactoryTable: syncMap.NewMap()}
}

func GetInstance() *MQClientManager {
	if instance == nil {
		once.Do(func() {
			instance = NewMQClientManager()
		})
	}
	return instance
}

// 从集合中查询MQClientInstance，无则创建一个
func (mQClientManager *MQClientManager) GetAndCreateMQClientInstance(clientConfig *stgclient.ClientConfig) *MQClientInstance {
	clientId := clientConfig.BuildMQClientId()
	instance, _ := mQClientManager.FactoryTable.Get(clientId)
	if nil == instance {
		instance = NewMQClientInstance(clientConfig.CloneClientConfig(), atomic.AddInt32(&mQClientManager.FactoryIndexGenerator, 1), clientId)
		prev, _ := mQClientManager.FactoryTable.PutIfAbsent(clientId, instance)
		if prev != nil {
			instance = prev
		} else {
			logger.Infof("first CreateMQClientInstance clientId=%v", clientId)
		}
	}
	return instance.(*MQClientInstance)
}

// 删除客户端
func (mQClientManager *MQClientManager) RemoveClientFactory(clientId string) {
	mQClientManager.FactoryTable.Remove(clientId)
}
