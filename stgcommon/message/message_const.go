package message

const (
	// 消息关键词，多个Key用KEY_SEPARATOR隔开（查询消息使用）
	PROPERTY_KEYS = "KEYS"

	// 消息标签，只支持设置一个Tag（服务端消息过滤使用）
	PROPERTY_TAGS = "TAGS"

	// 是否等待服务器将消息存储完毕再返回（可能是等待刷盘完成或者等待同步复制到其他服务器）
	PROPERTY_WAIT_STORE_MSG_OK = "WAIT"

	// 消息延时投递时间级别，0表示不延时，大于0表示特定延时级别（具体级别在服务器端定义）
	PROPERTY_DELAY_TIME_LEVEL = "DELAY"


	// 内部使用
	PROPERTY_RETRY_TOPIC = "RETRY_TOPIC"
	PROPERTY_REAL_TOPIC = "REAL_TOPIC"
	PROPERTY_REAL_QUEUE_ID = "REAL_QID"
	PROPERTY_TRANSACTION_PREPARED = "TRAN_MSG"
	PROPERTY_PRODUCER_GROUP = "PGROUP"
	PROPERTY_MIN_OFFSET = "MIN_OFFSET"
	PROPERTY_MAX_OFFSET = "MAX_OFFSET"
	PROPERTY_BUYER_ID = "BUYER_ID"
	PROPERTY_ORIGIN_MESSAGE_ID = "ORIGIN_MESSAGE_ID"
	PROPERTY_TRANSFER_FLAG = "TRANSFER_FLAG"
	PROPERTY_CORRECTION_FLAG = "CORRECTION_FLAG"
	PROPERTY_MQ2_FLAG = "MQ2_FLAG"
	PROPERTY_RECONSUME_TIME = "RECONSUME_TIME"
	KEY_SEPARATOR = " "
)
