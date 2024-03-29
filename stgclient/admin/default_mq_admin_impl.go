package admin

import (
	"bytes"
	"fmt"
	"git.oschina.net/cloudzone/smartgo/stgcommon"
	"git.oschina.net/cloudzone/smartgo/stgcommon/admin"
	"git.oschina.net/cloudzone/smartgo/stgcommon/logger"
	"git.oschina.net/cloudzone/smartgo/stgcommon/message"
	"git.oschina.net/cloudzone/smartgo/stgcommon/message/track"
	namesrvUtils "git.oschina.net/cloudzone/smartgo/stgcommon/namesrv"
	code "git.oschina.net/cloudzone/smartgo/stgcommon/protocol"
	"git.oschina.net/cloudzone/smartgo/stgcommon/protocol/body"
	"git.oschina.net/cloudzone/smartgo/stgcommon/protocol/heartbeat"
	"git.oschina.net/cloudzone/smartgo/stgcommon/protocol/route"
	"git.oschina.net/cloudzone/smartgo/stgcommon/subscription"
	"git.oschina.net/cloudzone/smartgo/stgcommon/utils"
	set "github.com/deckarep/golang-set"
	"strings"
)

const (
	timeoutMillis = int64(3 * 1000)
)

// 更新Broker配置
func (impl *DefaultMQAdminExtImpl) UpdateBrokerConfig(brokerAddr string, properties map[string]interface{}) error {
	// TODO
	return nil
}

// 创建或更新Topic
func (impl *DefaultMQAdminExtImpl) CreateAndUpdateTopicConfig(brokerAddr string, topicConfig *stgcommon.TopicConfig) error {
	return impl.mqClientInstance.MQClientAPIImpl.CreateTopic(brokerAddr, stgcommon.DEFAULT_TOPIC, topicConfig, int(timeoutMillis))
}

// 向指定Broker创建或者更新订阅组配置
func (impl *DefaultMQAdminExtImpl) CreateAndUpdateSubscriptionGroupConfig(addr string, config *subscription.SubscriptionGroupConfig) error {
	// TODO
	return nil
}

// 查询指定Broker的订阅组配置
func (impl *DefaultMQAdminExtImpl) ExamineSubscriptionGroupConfig(addr, group string) (*subscription.SubscriptionGroupConfig, error) {
	return nil, nil
}

// 查询指定Broker的Topic配置
func (impl *DefaultMQAdminExtImpl) ExamineTopicConfig(addr, topic string) (*stgcommon.TopicConfig, error) {
	return nil, nil
}

// 查询Topic Offset信息
func (impl *DefaultMQAdminExtImpl) ExamineTopicStats(topic string) (*admin.TopicStatsTable, error) {
	result := admin.NewTopicStatsTable()
	topicRouteData, err := impl.ExamineTopicRouteInfo(topic)
	if err != nil {
		return result, err
	}
	if topicRouteData == nil || topicRouteData.BrokerDatas == nil {
		return result, nil
	}
	for _, bd := range topicRouteData.BrokerDatas {
		brokerAddr := bd.SelectBrokerAddr()
		if brokerAddr != "" {
			tst, err := impl.mqClientInstance.MQClientAPIImpl.GetTopicStatsInfo(brokerAddr, topic, timeoutMillis)
			if err != nil {
				logger.Errorf("ExamineTopicStats err: %s, topic: %s, brokerAddr: %s", err.Error(), topic, brokerAddr)
				return result, err
			}

			if tst != nil && tst.OffsetTable != nil {
				for mq, topicOffset := range tst.OffsetTable {
					result.OffsetTable[mq] = topicOffset
				}
			}
		}
	}

	if len(result.OffsetTable) == 0 {
		return result, fmt.Errorf("not found the topic stats info")
	}
	return result, nil
}

// 从Name Server获取所有Topic列表
func (impl *DefaultMQAdminExtImpl) FetchAllTopicList() (*body.TopicList, error) {
	topicList, err := impl.mqClientInstance.MQClientAPIImpl.GetTopicListFromNameServer(timeoutMillis)
	return topicList, err
}

// GetTopicsByCluster 根据ClusterName，查询该集群管理的所有Topic
// Author: tianyuliang
// Since: 2017/11/8
func (impl *DefaultMQAdminExtImpl) GetTopicsByCluster(clusterName string) ([]*body.TopicBrokerClusterWapper, error) {
	topicBrokerClusterList := make([]*body.TopicBrokerClusterWapper, 0)

	topicList, err := impl.mqClientInstance.MQClientAPIImpl.GetTopicsByCluster(clusterName, timeoutMillis)
	if err != nil {
		return topicBrokerClusterList, err
	}
	if topicList == nil || len(topicList.TopicList) == 0 {
		return topicBrokerClusterList, fmt.Errorf("TopicPlusList.topics is empty. clusterName = %s", clusterName)
	}

	for _, tc := range topicList.TopicList {
		if topicList.TopicQueueTable != nil {
			for to, queueDatas := range topicList.TopicQueueTable {
				if queueDatas != nil {
					for _, queueData := range queueDatas {
						if queueData != nil && tc == to {
							topicBrokerCluster := body.NewTopicBrokerClusterWapper(clusterName, tc, queueData)
							topicBrokerClusterList = append(topicBrokerClusterList, topicBrokerCluster)
						}
					}
				}
			}
		}
	}
	return topicBrokerClusterList, nil
}

// 获取Broker运行时数据
func (impl *DefaultMQAdminExtImpl) FetchBrokerRuntimeStats(brokerAddr string) (*body.KVTable, error) {
	return impl.mqClientInstance.MQClientAPIImpl.GetBrokerRuntimeInfo(brokerAddr, timeoutMillis)
}

// 查询消费进度
func (impl *DefaultMQAdminExtImpl) ExamineConsumeStats(consumerGroup string) (*admin.ConsumeStats, error) {
	return impl.ExamineConsumeStatsByTopic(consumerGroup, "")
}

// 基于Topic查询消费进度
func (impl *DefaultMQAdminExtImpl) ExamineConsumeStatsByTopic(consumerGroup, topic string) (*admin.ConsumeStats, error) {
	result := admin.NewConsumeStats()

	retryTopic := stgcommon.GetRetryTopic(consumerGroup)
	if topic != "" {
		retryTopic = topic
	}
	topicRouteData, err := impl.ExamineTopicRouteInfo(retryTopic)
	if err != nil {
		return result, err
	}
	if topicRouteData == nil || topicRouteData.BrokerDatas == nil {
		return result, nil
	}
	for _, bd := range topicRouteData.BrokerDatas {
		brokerAddr := bd.SelectBrokerAddr()
		if brokerAddr != "" {
			big_timeoutMillis := int64(15 * 1000) // 由于查询时间戳会产生IO操作，可能会耗时较长，所以超时时间设置为15s
			consumeStats, err := impl.mqClientInstance.MQClientAPIImpl.GetConsumeStatsByTopic(brokerAddr, consumerGroup, topic, big_timeoutMillis)
			if err != nil {
				return result, err
			}
			if consumeStats != nil && consumeStats.OffsetTable != nil {
				for key, value := range consumeStats.OffsetTable {
					result.OffsetTable[key] = value
				}
			}
			result.ConsumeTps += consumeStats.ConsumeTps
		}
	}
	if len(result.OffsetTable) == 0 {
		format := "not found the consumer group consume stats, because return offset table is empty, maybe the consumer not consume any message"
		return result, fmt.Errorf(format)
	}
	return result, nil
}

// 查看集群信息
func (impl *DefaultMQAdminExtImpl) ExamineBrokerClusterInfo() (*body.ClusterPlusInfo, []*body.ClusterBrokerWapper, error) {
	return impl.mqClientInstance.MQClientAPIImpl.GetBrokerClusterInfo(timeoutMillis)
}

// 查看Topic路由信息
func (impl *DefaultMQAdminExtImpl) ExamineTopicRouteInfo(topic string) (*route.TopicRouteData, error) {
	return impl.mqClientInstance.MQClientAPIImpl.GetTopicRouteInfoFromNameServer(topic, timeoutMillis)
}

// 查看Consumer网络连接、订阅关系
func (impl *DefaultMQAdminExtImpl) ExamineConsumerConnectionInfo(consumerGroup, topic string) (*body.ConsumerConnectionPlus, int, error) {
	result := body.NewConsumerConnectionPlus()
	onlineCode := 0
	retryTopic := stgcommon.GetRetryTopic(consumerGroup)
	if topic != "" {
		retryTopic = topic
	}
	topicRouteData, err := impl.ExamineTopicRouteInfo(retryTopic)
	if err != nil {
		return result, onlineCode, err
	}
	if topicRouteData == nil || topicRouteData.BrokerDatas == nil {
		return result, onlineCode, nil
	}
	for _, bd := range topicRouteData.BrokerDatas {
		brokerAddr := bd.SelectBrokerAddr()
		if brokerAddr != "" {
			return impl.mqClientInstance.MQClientAPIImpl.GetConsumerConnectionList(brokerAddr, consumerGroup, timeoutMillis)
		}
	}

	return result, onlineCode, fmt.Errorf("not found the consumer group connection")
}

// 查看Producer网络连接
func (impl *DefaultMQAdminExtImpl) ExamineProducerConnectionInfo(producerGroup, topic string) (*body.ProducerConnection, error) {
	result := body.NewProducerConnection()
	topicRouteData, err := impl.ExamineTopicRouteInfo(topic)
	if err != nil {
		return result, err
	}
	if topicRouteData == nil || topicRouteData.BrokerDatas == nil {
		return result, nil
	}
	for _, bd := range topicRouteData.BrokerDatas {
		brokerAddr := bd.SelectBrokerAddr()
		if brokerAddr != "" {
			return impl.mqClientInstance.MQClientAPIImpl.GetProducerConnectionList(brokerAddr, producerGroup, timeoutMillis)
		}
	}

	format := "not found the producer group connection"
	return result, fmt.Errorf(format)
}

// 获取Name Server地址列表
func (impl *DefaultMQAdminExtImpl) GetNameServerAddressList() ([]string, error) {
	return impl.mqClientInstance.MQClientAPIImpl.GetNameServerAddressList(), nil
}

// 清除某个Broker的写权限，针对所有Name Server
// return 返回清除了多少个topic
func (impl *DefaultMQAdminExtImpl) WipeWritePermOfBroker(namesrvAddr, brokerName string) (int, error) {
	return impl.mqClientInstance.MQClientAPIImpl.WipeWritePermOfBroker(namesrvAddr, brokerName, timeoutMillis)
}

// 向Name Server增加一个配置项
func (impl *DefaultMQAdminExtImpl) PutKVConfig(namespace, key, value string) error {
	return impl.mqClientInstance.MQClientAPIImpl.PutKVConfigValue(namespace, key, value, timeoutMillis)
}

// 从Name Server获取一个配置项
func (impl *DefaultMQAdminExtImpl) GetKVConfig(namespace, key string) (string, error) {
	return impl.mqClientInstance.MQClientAPIImpl.GetKVConfigValue(namespace, key, timeoutMillis)
}

// 在 namespace 上添加或者更新 KV 配置
func (impl *DefaultMQAdminExtImpl) CreateAndUpdateKvConfig(namespace, key, value string) error {
	return impl.mqClientInstance.MQClientAPIImpl.PutKVConfigValue(namespace, key, value, timeoutMillis)
}

// 删除 namespace 上的 KV 配置
func (impl *DefaultMQAdminExtImpl) DeleteKvConfig(namespace, key string) error {
	// TODO
	return nil
}

// 获取指定Namespace下的所有kv
func (impl *DefaultMQAdminExtImpl) GetKVListByNamespace(namespace string) (*body.KVTable, error) {
	return impl.mqClientInstance.MQClientAPIImpl.GetKVListByNamespace(namespace, timeoutMillis)
}

// 删除 broker 上的 topic 信息
func (impl *DefaultMQAdminExtImpl) DeleteTopicInBroker(brokerAddrs set.Set, topic string) error {
	if brokerAddrs == nil {
		brokerAddrs = set.NewSet()
	}
	for brokerAddr := range brokerAddrs.Iterator().C {
		err := impl.mqClientInstance.MQClientAPIImpl.DeleteTopicInBroker(brokerAddr.(string), topic, timeoutMillis)
		if err != nil {
			logger.Errorf("delete topic from broker err: %s", err.Error())
		} else {
			logger.Warnf("delete topic from broker ok. brokerAddr: %s, topic: %s", brokerAddr.(string), topic)
		}
	}
	return nil
}

// 删除 namesrv维护的topic信息
func (impl *DefaultMQAdminExtImpl) DeleteTopicInNameServer(namesrvSet set.Set, topic string) error {
	if namesrvSet == nil {
		namesrvSet = set.NewSet()
		// TODO 发送HTTP请求，最终调用TopAddressing.fetchNSAddr()获取最新的namesrv地址
	}
	for namesrvAddr := range namesrvSet.Iterator().C {
		err := impl.mqClientInstance.MQClientAPIImpl.DeleteTopicInNameServer(namesrvAddr.(string), topic, timeoutMillis)
		if err != nil {
			logger.Errorf("delete topic from namesrv err: %s", err.Error())
		} else {
			logger.Warnf("delete topic from namesrv ok. namesrvAddr: %s, topic: %s", namesrvAddr.(string), topic)
		}
	}
	return nil
}

// 删除 broker 上的 subscription group 信息
func (impl *DefaultMQAdminExtImpl) DeleteSubscriptionGroup(brokerAddr, groupName string) error {
	return impl.mqClientInstance.MQClientAPIImpl.DeleteSubscriptionGroup(brokerAddr, groupName, timeoutMillis)
}

// 通过 server ip 获取 project 信息
func (impl *DefaultMQAdminExtImpl) GetProjectGroupByIp(ip string) (string, error) {
	projectGroup, err := impl.mqClientInstance.MQClientAPIImpl.GetProjectGroupByIp(ip, timeoutMillis)
	return projectGroup, err
}

// 通过 project 获取所有的 server ip 信息
func (impl *DefaultMQAdminExtImpl) GetIpsByProjectGroup(projectGroup string) (string, error) {
	// TODO
	return "", nil
}

// 删除 project group 对应的所有 server ip
func (impl *DefaultMQAdminExtImpl) DeleteIpsByProjectGroup(key string) error {
	// TODO
	return nil
}

// 按照时间回溯消费进度(客户端需要重启)
func (impl *DefaultMQAdminExtImpl) ResetOffsetByTimestampOld(consumerGroup, topic string, timestamp int64, force bool) ([]*admin.RollbackStats, error) {
	// TODO
	return nil, nil
}

// 按照时间回溯消费进度(客户端不需要重启)
func (impl *DefaultMQAdminExtImpl) ResetOffsetByTimestamp(topic, group string, timestamp int64, force bool) (map[*message.MessageQueue]int64, error) {
	// TODO
	return nil, nil
}

// 重置消费进度，无论Consumer是否在线，都可以执行。不保证最终结果是否成功，需要调用方通过消费进度查询来再次确认
func (impl *DefaultMQAdminExtImpl) ResetOffsetNew(consumerGroup, topic string, timestamp int64) error {
	// TODO
	return nil
}

// 通过客户端查看消费者的消费情况
func (impl *DefaultMQAdminExtImpl) GetConsumeStatus(topic, consumerGroupId, clientAddr string) (map[string]map[*message.MessageQueue]int64, error) {
	result := make(map[string]map[*message.MessageQueue]int64)
	topicRouteData, err := impl.ExamineTopicRouteInfo(topic)
	if err != nil {
		return result, err
	}
	if topicRouteData == nil || topicRouteData.BrokerDatas == nil {
		return result, nil
	}
	for _, bd := range topicRouteData.BrokerDatas {
		// 每个 broker 上有所有的 consumer 连接，故只需要在一个 broker 执行即可
		brokerAddr := bd.SelectBrokerAddr()
		if brokerAddr != "" {
			impl.mqClientInstance.MQClientAPIImpl.InvokeBrokerToGetConsumerStatus(brokerAddr, topic, consumerGroupId, clientAddr, 5000)
		}
	}
	return result, nil

}

// 创建或更新顺序消息的分区配置
func (impl *DefaultMQAdminExtImpl) CreateOrUpdateOrderConf(key, value string, isCluster bool) error {
	if isCluster {
		impl.mqClientInstance.MQClientAPIImpl.PutKVConfigValue(namesrvUtils.NAMESPACE_ORDER_TOPIC_CONFIG, key, value, timeoutMillis)
		return nil
	}

	oldOrderConfs, err := impl.mqClientInstance.MQClientAPIImpl.GetKVConfigValue(namesrvUtils.NAMESPACE_ORDER_TOPIC_CONFIG, key, timeoutMillis)
	if err != nil {
		logger.Errorf("CreateOrUpdateOrderConf err: %s", err.Error())
		return err
	}

	orderConfMap := make(map[string]string)
	if !stgcommon.IsEmpty(oldOrderConfs) {
		oldOrderConfArr := strings.Split(oldOrderConfs, ";")
		for _, oldOrderConf := range oldOrderConfArr {
			items := strings.Split(oldOrderConf, ":")
			orderConfMap[items[0]] = oldOrderConf
		}
	}

	items := strings.Split(value, ":")
	orderConfMap[items[0]] = value

	newOrderConf := &bytes.Buffer{}
	splitor := ""
	for _, value := range orderConfMap {
		newOrderConf.WriteString(splitor)
		newOrderConf.WriteString(value)
		splitor = ";"
	}

	return impl.mqClientInstance.MQClientAPIImpl.PutKVConfigValue(namesrvUtils.NAMESPACE_ORDER_TOPIC_CONFIG, key, newOrderConf.String(), timeoutMillis)
}

// 根据Topic查询被哪些订阅组消费
func (impl *DefaultMQAdminExtImpl) QueryTopicConsumeByWho(topic string) (*body.GroupList, error) {
	topicRouteData, err := impl.ExamineTopicRouteInfo(topic)
	if err != nil {
		return nil, err
	}
	if topicRouteData == nil || topicRouteData.BrokerDatas == nil {
		return nil, err
	}
	for _, bd := range topicRouteData.BrokerDatas {
		brokerAddr := bd.SelectBrokerAddr()
		if brokerAddr != "" {
			return impl.mqClientInstance.MQClientAPIImpl.QueryTopicConsumeByWho(brokerAddr, topic, timeoutMillis)
		}
		break
	}
	return nil, fmt.Errorf("QueryTopicConsumeByWho failed, but unknown reason. topic = %s", topic)
}

// 根据 topic 和 group 获取消息的时间跨度
// retutn set<QueueTimeSpan>
func (impl *DefaultMQAdminExtImpl) QueryConsumeTimeSpan(topic, consumerGroupId string) (set.Set, error) {
	spanSet := set.NewSet()
	topicRouteData, err := impl.ExamineTopicRouteInfo(topic)
	if err != nil {
		return spanSet, err
	}
	if topicRouteData == nil || topicRouteData.BrokerDatas == nil {
		return spanSet, nil
	}
	for _, bd := range topicRouteData.BrokerDatas {
		brokerAddr := bd.SelectBrokerAddr()
		if brokerAddr == "" {
			continue
		}
		qcts, err := impl.mqClientInstance.MQClientAPIImpl.QueryConsumeTimeSpan(brokerAddr, topic, consumerGroupId, timeoutMillis)
		if err != nil {
			logger.Errorf("QueryConsumeTimeSpan err: %s", err.Error())
			continue
		}
		spanSet = spanSet.Union(qcts)
	}
	return spanSet, nil
}

// 触发清理失效的消费队列
// cluster 如果参数cluster为空，则表示所有集群
// return 清理是否成功
func (impl *DefaultMQAdminExtImpl) CleanExpiredConsumerQueue(clusterName string) (result bool, err error) {
	clusterPlusInfo, _, err := impl.ExamineBrokerClusterInfo()
	if err != nil {
		return false, err
	}
	if clusterName == "" {
		if clusterPlusInfo == nil || clusterPlusInfo.ClusterAddrTable == nil {
			return false, nil
		}
		for targetCluster, _ := range clusterPlusInfo.ClusterAddrTable {
			result, err = impl.cleanExpiredConsumerQueueByCluster(clusterPlusInfo, targetCluster)
		}
	} else {
		result, err = impl.cleanExpiredConsumerQueueByCluster(clusterPlusInfo, clusterName)
	}
	return result, err
}

// cleanExpiredConsumerQueueByCluster 根据集群名称，清除过期的消费队列
// Author: tianyuliang
// Since: 2017/11/6
func (impl *DefaultMQAdminExtImpl) cleanExpiredConsumerQueueByCluster(clusterInfo *body.ClusterPlusInfo, clusterName string) (result bool, err error) {
	if clusterInfo == nil {
		return false, nil
	}

	brokerAddrs, _ := clusterInfo.RetrieveAllAddrByCluster(clusterName)
	for _, brokerAddr := range brokerAddrs {
		result, err = impl.CleanExpiredConsumerQueueByAddr(brokerAddr)
	}
	return result, err
}

// 触发指定的broker清理失效的消费队列
// return 清理是否成功
func (impl *DefaultMQAdminExtImpl) CleanExpiredConsumerQueueByAddr(brokerAddr string) (bool, error) {
	result, err := impl.mqClientInstance.MQClientAPIImpl.CleanExpiredConsumeQueue(brokerAddr, timeoutMillis)
	if err != nil {
		format := "clean expired ConsumeQueue on target broker[%s] err: %s"
		logger.Infof(format, brokerAddr, err.Error())
		return result, err
	}
	format := "clean expired ConsumeQueue on target broker[%s], the result is [%t]"
	logger.Infof(format, brokerAddr, result)
	return result, nil
}

// 查询Consumer内存数据结构
func (impl *DefaultMQAdminExtImpl) GetConsumerRunningInfo(consumerGroupId, clientId string, jstack bool) (*body.ConsumerRunningInfo, error) {
	consumerRunningInfo := body.NewConsumerRunningInfo()
	retryTopic := stgcommon.RETRY_GROUP_TOPIC_PREFIX + consumerGroupId
	topicRouteData, err := impl.ExamineTopicRouteInfo(retryTopic)
	if err != nil {
		return nil, err
	}
	if topicRouteData == nil || topicRouteData.BrokerDatas == nil {
		return consumerRunningInfo, nil
	}
	big_timeoutMills := int64(12 * 1000)
	for _, bd := range topicRouteData.BrokerDatas {
		brokerAddr := bd.SelectBrokerAddr()
		if brokerAddr != "" {
			return impl.mqClientInstance.MQClientAPIImpl.GetConsumerRunningInfo(brokerAddr, consumerGroupId, clientId, jstack, big_timeoutMills)
		}
	}
	return consumerRunningInfo, err
}

// 向指定Consumer发送某条消息
func (impl *DefaultMQAdminExtImpl) ConsumeMessageDirectly(consumerGroup, clientId, msgId string) (*body.ConsumeMessageDirectlyResult, error) {
	msg, err := impl.ViewMessage(msgId)
	if err != nil {
		return nil, err
	}
	return impl.mqClientInstance.MQClientAPIImpl.ConsumeMessageDirectly(msg.StoreHost, consumerGroup, clientId, msgId, timeoutMillis)
}

//查询消息被谁消费了
func (impl *DefaultMQAdminExtImpl) MessageTrackDetail(msg *message.MessageExt) ([]*track.MessageTrack, error) {
	result := make([]*track.MessageTrack, 0)
	groupList, err := impl.QueryTopicConsumeByWho(msg.Topic)
	if err != nil {
		logger.Errorf("DefaultMQAdminExtImpl.QueryTopicConsumeByWho() err: %s, topic: %s", err.Error(), msg.Topic)
		return result, err
	}
	if groupList == nil || groupList.GroupList == nil || groupList.GroupList.Cardinality() == 0 {
		//TODO: broker没有返回GroupList列表, 那么默认“MessageTrack”：未订阅，未消费
		msgTrack := new(track.MessageTrack)
		msgTrack.TrackType = track.NotSubscribedAndNotConsumed
		msgTrack.Code = code.SUCCESS
		msgTrack.ExceptionDesc = ""

		result = append(result, msgTrack)
		return result, nil
	}
	var tracks []*track.MessageTrack
	for itor := range groupList.GroupList.Iterator().C {
		if consumerGroupId, ok := itor.(string); ok {
			messageTrack := track.NewMessageTrack(consumerGroupId)
			cc, onlineCode, err := impl.ExamineConsumerConnectionInfo(consumerGroupId, msg.Topic)
			if err != nil {
				if onlineCode == code.CONSUMER_NOT_ONLINE {
					messageTrack.TrackType = track.ConsumerGroupIdNotOnline
					messageTrack.ExceptionDesc = ""
				} else {
					messageTrack.TrackType = track.UnknowExeption
					messageTrack.ExceptionDesc = err.Error()
				}
				messageTrack.Code = code.SYSTEM_ERROR
				tracks = append(tracks, messageTrack)
				continue
			}
			if cc == nil || cc.ConnectionSet == nil || len(cc.ConnectionSet) == 0 {
				messageTrack.Code = code.CONSUMER_NOT_ONLINE
				messageTrack.ExceptionDesc = fmt.Sprintf("the consumer group[%s] not online.", consumerGroupId)
				tracks = append(tracks, messageTrack)
				continue
			}

			switch cc.ConsumeType {
			case heartbeat.CONSUME_ACTIVELY:
				messageTrack.TrackType = track.SubscribedButPull
				messageTrack.Code = code.SUCCESS
			case heartbeat.CONSUME_PASSIVELY:
				flag, err := impl.Consumed(msg, consumerGroupId)
				if err != nil {
					messageTrack.Code = code.SYSTEM_ERROR
					messageTrack.ExceptionDesc = err.Error()
					break
				}

				if flag {
					messageTrack.TrackType = track.SubscribedAndConsumed
					messageTrack.Code = code.SUCCESS
					// 查看订阅关系是否匹配
					if cc != nil && cc.SubscriptionTable != nil {
						for topic, subscriptionData := range cc.SubscriptionTable {
							if topic != msg.Topic {
								continue
							}
							if subscriptionData == nil || subscriptionData.TagsSet == nil {
								continue
							}
							for _, msgTag := range subscriptionData.TagsSet {
								if msgTag != msg.GetTags() && msgTag != "*" {
									messageTrack.TrackType = track.SubscribedButFilterd
									messageTrack.Code = code.SUCCESS
								}
							}
						}
					}
				} else {
					messageTrack.TrackType = track.SubscribedAndNotConsumeYet
					messageTrack.Code = code.SUCCESS
				}
			default:
			}
			tracks = append(tracks, messageTrack)
		}
	}

	return tracks, nil
}

// Consumed 校验某条消息是否被某个消费组消费过
//
// return: true表示已被消费； false:表示未被消费
//
// Author: tianyuliang
// Since: 2017/11/6
func (impl *DefaultMQAdminExtImpl) Consumed(msg *message.MessageExt, consumerGroupId string) (bool, error) {
	ci, _, err := impl.ExamineBrokerClusterInfo()
	if err != nil {
		return false, err
	}
	if ci.BrokerAddrTable == nil || len(ci.BrokerAddrTable) == 0 {
		return false, nil
	}

	cstats, err := impl.ExamineConsumeStats(consumerGroupId)
	if err != nil {
		return false, err
	}
	if cstats.OffsetTable == nil || len(cstats.OffsetTable) == 0 {
		return false, nil
	}

	for mq, offsetwapper := range cstats.OffsetTable {
		if mq != nil && mq.Topic == msg.Topic && int32(mq.QueueId) == msg.QueueId {
			if brokerData, ok := ci.BrokerAddrTable[mq.BrokerName]; ok && brokerData != nil {
				if brokerAddr, ok := brokerData.BrokerAddrs[stgcommon.MASTER_ID]; ok && brokerAddr != "" {
					format := "brokerAddr=%s, msg.StoreHost=%s, offsetwapper.ConsumerOffset=%d, msg.QueueOffset=%d"
					logger.Infof(format, brokerAddr, msg.StoreHost, offsetwapper.ConsumerOffset, msg.QueueOffset)

					if brokerAddr == msg.StoreHost {
						if offsetwapper != nil && offsetwapper.ConsumerOffset > msg.QueueOffset {
							return true, nil
						}
					}
				}
			}
		}
	}
	return false, nil
}

// 克隆某一个组的消费进度到新的组
func (impl *DefaultMQAdminExtImpl) CloneGroupOffset(srcGroup, destGroup, topic string, isOffline bool) error {
	retryTopic := stgcommon.GetRetryTopic(srcGroup)
	topicRouteData, err := impl.ExamineTopicRouteInfo(retryTopic)
	if err != nil {
		return err
	}
	if topicRouteData == nil || topicRouteData.BrokerDatas == nil {
		return nil
	}
	for _, bd := range topicRouteData.BrokerDatas {
		brokerAddr := bd.SelectBrokerAddr()
		if brokerAddr == "" {
			continue
		}
		impl.mqClientInstance.MQClientAPIImpl.CloneGroupOffset(brokerAddr, srcGroup, destGroup, topic, isOffline, timeoutMillis)
	}
	return nil
}

// 服务器统计数据输出
func (impl *DefaultMQAdminExtImpl) ViewBrokerStatsData(brokerAddr, statsName, statsKey string) (*body.BrokerStatsData, error) {
	return impl.mqClientInstance.MQClientAPIImpl.ViewBrokerStatsData(brokerAddr, statsName, statsKey, timeoutMillis)
}

// 创建Topic
// key 消息队列已存在的topic
// newTopic 需新建的topic
// queueNum 读写队列的数量
func (impl *DefaultMQAdminExtImpl) CreateTopic(key, newTopic string, queueNum int) error {
	return impl.mqClientInstance.MQAdminImpl.CreateTopic(key, newTopic, queueNum, 0)
}

// 创建Topic
// key 消息队列已存在的topic
// newTopic 需新建的topic
// queueNum 读写队列的数量
func (impl *DefaultMQAdminExtImpl) CreateCustomTopic(brokerAddr string, topicConfig *stgcommon.TopicConfig) error {
	return impl.mqClientInstance.MQClientAPIImpl.CreateTopic(brokerAddr, stgcommon.DEFAULT_TOPIC, topicConfig, int(timeoutMillis))
}

// 根据msgId查询消息消费结果
func (impl *DefaultMQAdminExtImpl) ViewMessage(msgId string) (*message.MessageExt, error) {
	messageId, err := message.DecodeMessageId(msgId)
	if err != nil {
		return nil, err
	}

	storeHost := messageId.Address      // 消息存储的实际broker地址
	commitLogOffset := messageId.Offset // 消息存储的物理偏移量
	return impl.mqClientInstance.MQClientAPIImpl.ViewMessage(storeHost, commitLogOffset, timeoutMillis)
}

// 搜索消息
// topic  topic名称
// key    消息key关键字[业务系统基于此字段唯一标识消息]
// maxNum 最大搜索条数
// begin  开始查询消息的时间戳
// end    结束查询消息的时间戳
func (impl *DefaultMQAdminExtImpl) QueryMessage(topic, key string, maxNum int, begin, end int64) (*admin.QueryResult, error) {
	return nil, nil
}

// 查询较早的存储消息
func (impl *DefaultMQAdminExtImpl) EarliestMsgStoreTime(mq *message.MessageQueue) (int64, error) {
	return 0, nil
}

// 根据时间戳搜索MessageQueue偏移量(注意:可能会出现大量IO开销)
func (impl *DefaultMQAdminExtImpl) SearchOffset(mq message.MessageQueue, timestamp int64) (int64, error) {
	return 0, nil
}

// 查询MessageQueue最大偏移量
func (impl *DefaultMQAdminExtImpl) MaxOffset(mq *message.MessageQueue) (int64, error) {
	return impl.mqClientInstance.MQAdminImpl.MaxOffset(mq), nil
}

// 查询MessageQueue最小偏移量
func (impl *DefaultMQAdminExtImpl) MinOffset(mq *message.MessageQueue) (int64, error) {
	return 0, nil
}

// FetchMasterAddrByClusterName 拉取所有角色是“master”的broker地址列表
//
// 返回值: set.Set保存所有角色是master的 brokerAddr地址,即set<brokerAddr>
//
// Author: tianyuliang
// Since: 2017/11/7
func (impl *DefaultMQAdminExtImpl) FetchMasterAddrByClusterName(clusterName string) (set.Set, error) {
	masterSet := set.NewSet()
	clusterInfoWrapper, _, err := impl.ExamineBrokerClusterInfo()
	if err != nil {
		return masterSet, err
	}
	if clusterInfoWrapper == nil || clusterInfoWrapper.ClusterAddrTable == nil {
		return masterSet, nil
	}

	brokerNameSet, ok := clusterInfoWrapper.ClusterAddrTable[clusterName]
	if !ok || brokerNameSet == nil || len(brokerNameSet) == 0 {
		logger.Error("[error] Make sure the specified clusterName exists or the nameserver which connected is correct.")
		return masterSet, nil
	}
	for _, brokerName := range brokerNameSet {
		brokerData, ok := clusterInfoWrapper.BrokerAddrTable[brokerName]
		if ok && brokerData != nil && brokerData.BrokerAddrs != nil {
			brokerAddr := brokerData.BrokerAddrs[stgcommon.MASTER_ID]
			if brokerAddr != "" {
				masterSet.Add(brokerAddr)
			}
		}
	}
	return masterSet, nil
}

// FetchBrokerNameByClusterName 根据Cluster集群名称，拉取所有broker名称
//
// 返回值: set<brokerName>
//
// Author: tianyuliang
// Since: 2017/11/7
func (impl *DefaultMQAdminExtImpl) FetchBrokerNameByClusterName(clusterName string) (set.Set, error) {
	clusterInfoWrapper, _, err := impl.ExamineBrokerClusterInfo()
	if err != nil {
		return nil, err
	}
	if clusterInfoWrapper != nil && clusterInfoWrapper.ClusterAddrTable != nil {
		brokerNames, ok := clusterInfoWrapper.ClusterAddrTable[clusterName]
		if ok && len(brokerNames) > 0 {
			brokerNameSet := set.NewSet()
			for _, bn := range brokerNames {
				brokerNameSet.Add(bn)
			}
			return brokerNameSet, nil
		}
	}

	format := "Make sure the specified clusterName exists or the nameserver which connected is correct."
	return nil, fmt.Errorf(format)
}

// FetchBrokerNameByAddr 根据broker地址查询对应的broker名称
//
// 返回值: set<brokerName>
//
// Author: tianyuliang
// Since: 2017/11/7
func (impl *DefaultMQAdminExtImpl) FetchBrokerNameByAddr(brokerAddr string) (string, error) {
	clusterInfoWrapper, _, err := impl.ExamineBrokerClusterInfo()
	if err != nil {
		return "", err
	}
	if clusterInfoWrapper != nil && clusterInfoWrapper.BrokerAddrTable != nil {
		for brokerName, brokerData := range clusterInfoWrapper.BrokerAddrTable {
			if brokerData != nil && brokerData.BrokerAddrs != nil {
				for _, addr := range brokerData.BrokerAddrs {
					if strings.Contains(addr, brokerAddr) {
						return brokerName, nil
					}
				}
			}
		}
	}

	format := "Make sure the specified broker addr exists or the nameserver which connected is correct."
	return "", fmt.Errorf(format)
}

// GetClusterList 获取集群名称
// Author: tianyuliang
// Since: 2017/11/7
func (impl *DefaultMQAdminExtImpl) GetAllClusterNames() ([]string, map[string]*route.BrokerData, error) {
	clusterPlus, _, err := impl.ExamineBrokerClusterInfo()
	if err != nil {
		return []string{}, nil, err
	}
	if clusterPlus == nil || clusterPlus.ClusterAddrTable == nil || len(clusterPlus.ClusterAddrTable) == 0 {
		return []string{}, nil, fmt.Errorf("clusterInfoWrapper is nil, or clusterInfoWrapper.ClusterAddrTable is empty")
	}

	clusterNames := make([]string, 0, len(clusterPlus.ClusterAddrTable))
	brokerAddrTable := clusterPlus.BrokerAddrTable
	if brokerAddrTable == nil {
		brokerAddrTable = make(map[string]*route.BrokerData)
	}

	for clusterName, _ := range clusterPlus.ClusterAddrTable {
		clusterNames = append(clusterNames, clusterName)
	}
	return clusterNames, brokerAddrTable, nil
}

// GetClusterList 获取集群名称
// Author: tianyuliang
// Since: 2017/11/7
func (impl *DefaultMQAdminExtImpl) GetClusterTopicWappers() ([]*body.TopicBrokerClusterWapper, error) {
	defer utils.RecoveredFn()

	result := make([]*body.TopicBrokerClusterWapper, 0)
	clusterNames, brokerAddrTable, err := impl.GetAllClusterNames()
	if err != nil {
		return result, err
	}
	for _, clusterName := range clusterNames {
		topicBrokerClusterList, err := impl.GetTopicsByCluster(clusterName)
		if err != nil {
			return result, err
		}

		for _, topicBrokerCluster := range topicBrokerClusterList {
			brokerId, brokerAddr := impl.getBrokerAddrByName(brokerAddrTable, topicBrokerCluster.TopicUpdateConfigWapper.BrokerName)
			topicBrokerCluster.TopicUpdateConfigWapper.BrokerAddr = brokerAddr
			topicBrokerCluster.TopicUpdateConfigWapper.BrokerId = brokerId

			result = append(result, topicBrokerCluster)
		}
	}
	return result, nil
}

// getBrokerByName 查询brokerAddr地址
// Author: tianyuliang
// Since: 2017/11/8
func (impl *DefaultMQAdminExtImpl) getBrokerAddrByName(brokerAddrTable map[string]*route.BrokerData, brokerName string) (int, string) {
	if brokerAddrTable == nil || len(brokerAddrTable) == 0 {
		return 0, ""
	}
	for _, brokerData := range brokerAddrTable {
		if brokerData != nil && brokerData.BrokerAddrs != nil {
			for brokerId, brokerAddr := range brokerData.BrokerAddrs {
				if brokerId == stgcommon.MASTER_ID && brokerData.BrokerName == brokerName {
					return brokerId, brokerAddr
				}
			}
		}
	}
	return 0, ""
}
