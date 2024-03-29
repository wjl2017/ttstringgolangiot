package process

// 客户端对外使用的producer接口
// Author: yintongqiang
// Since:  2017/8/8

import "git.oschina.net/cloudzone/smartgo/stgcommon/message"

type MQProducer interface {
	// 启动
	Start() error
	// 关闭
	Shutdown()
	// 同步发送消息
	Send(msg *message.Message) (*SendResult, error)
	// 只发送不处理
	SendOneWay(msg *message.Message) error
	// 异步发送
	SendCallBack(msg *message.Message,callback SendCallback) error
}
