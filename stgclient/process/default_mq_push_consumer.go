package process

import (
	"git.oschina.net/cloudzone/smartgo/stgclient"
	"git.oschina.net/cloudzone/smartgo/stgclient/consumer/listener"
	"git.oschina.net/cloudzone/smartgo/stgclient/consumer/rebalance"
	"git.oschina.net/cloudzone/smartgo/stgclient/consumer/store"
	"git.oschina.net/cloudzone/smartgo/stgcommon/protocol/heartbeat"
)

// DefaultMQPushConsumer: push消费
// Author: yintongqiang
// Since:  2017/8/10

type DefaultMQPushConsumer struct {
	defaultMQPushConsumerImpl *DefaultMQPushConsumerImpl
	// Do the same thing for the same Group, the application must be set,and
	// guarantee Globally unique
	consumerGroup string
	// Consumption pattern,default is clustering
	messageModel heartbeat.MessageModel
	// Consumption offset
	consumeFromWhere heartbeat.ConsumeFromWhere
	// Backtracking consumption time with second precision.time format is
	// 20131223171201<br>
	// Implying Seventeen twelve and 01 seconds on December 23, 2013 year<br>
	// Default backtracking consumption time Half an hour ago
	consumeTimestamp string
	// Queue allocation algorithm
	allocateMessageQueueStrategy rebalance.AllocateMessageQueueStrategy
	// Subscription relationship
	subscription map[string]string
	// Message listener
	messageListener listener.MessageListener
	// Offset Storage
	offsetStore store.OffsetStore
	// Minimum consumer thread number
	consumeThreadMin int
	// Max consumer thread number
	consumeThreadMax int
	// Threshold for dynamic adjustment of the number of thread pool
	adjustThreadPoolNumsThreshold int64
	// Concurrently max span offset.it has no effect on sequential consumption
	consumeConcurrentlyMaxSpan int64
	// Flow control threshold
	pullThresholdForQueue int64
	// Message pull Interval
	pullInterval int64
	// Batch consumption size
	consumeMessageBatchMaxSize int
	// Batch pull size
	pullBatchSize int
	// Whether update subscription relationship when every pull
	postSubscriptionWhenPull bool
	// Whether the unit of subscription group
	unitMode     bool
	clientConfig *stgclient.ClientConfig
}

// 创建push消费结构体
func NewDefaultMQPushConsumer(consumerGroup string) *DefaultMQPushConsumer {
	pushConsumer := &DefaultMQPushConsumer{clientConfig: stgclient.NewClientConfig("")}
	pushConsumer.consumerGroup = consumerGroup
	pushConsumer.messageModel = heartbeat.CLUSTERING
	pushConsumer.pullThresholdForQueue = 1000
	pushConsumer.consumeMessageBatchMaxSize = 1
	pushConsumer.consumeThreadMax = 64
	pushConsumer.pullInterval = 0
	pushConsumer.pullBatchSize = 32
	pushConsumer.unitMode = false
	pushConsumer.consumeConcurrentlyMaxSpan = 2000
	pushConsumer.allocateMessageQueueStrategy = rebalance.AllocateMessageQueueAveragely{}
	pushConsumer.defaultMQPushConsumerImpl = NewDefaultMQPushConsumerImpl(pushConsumer)
	return pushConsumer
}

// 设置从哪个位置开始消费
func (pushConsumer *DefaultMQPushConsumer) SetConsumeFromWhere(consumeFromWhere heartbeat.ConsumeFromWhere) {
	pushConsumer.consumeFromWhere = consumeFromWhere
}

// 设置namesrvaddr
func (pushConsumer *DefaultMQPushConsumer) SetNamesrvAddr(namesrvAddr string) {
	pushConsumer.clientConfig.NamesrvAddr = namesrvAddr
}

// 设置消费类型
func (pushConsumer *DefaultMQPushConsumer) SetMessageModel(model heartbeat.MessageModel) {
	pushConsumer.messageModel = model
}

// 订阅topic和tag
func (pushConsumer *DefaultMQPushConsumer) Subscribe(topic string, subExpression string) {
	pushConsumer.defaultMQPushConsumerImpl.subscribe(topic, subExpression)
}

// 注册监听器
func (pushConsumer *DefaultMQPushConsumer) RegisterMessageListener(messageListener listener.MessageListener) {
	pushConsumer.messageListener = messageListener
	pushConsumer.defaultMQPushConsumerImpl.registerMessageListener(messageListener)
}

// 启动消费服务
func (pushConsumer *DefaultMQPushConsumer) Start() {
	pushConsumer.defaultMQPushConsumerImpl.Start()
}

// 关闭消费服务
func (pushConsumer *DefaultMQPushConsumer) Shutdown() {
	pushConsumer.defaultMQPushConsumerImpl.Shutdown()
}
