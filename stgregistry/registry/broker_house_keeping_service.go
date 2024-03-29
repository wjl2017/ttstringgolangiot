package registry

import (
	"git.oschina.net/cloudzone/smartgo/stgnet/netm"
	"git.oschina.net/cloudzone/smartgo/stgregistry/logger"
)

// BrokerHousekeepingServices Broker活动检测服务
//
// (1)ContextListener是smartnet模块封装接口
// (2)NameSrv监测Broker的死亡：当Broker和NameSrv之间的长连接断掉之后，回调ContextListener对应的函数，从而触发NameServer的路由信息更新
//
// Author: tianyuliang
// Since: 2017/9/6
type BrokerHousekeepingService struct {
	NamesrvController *DefaultNamesrvController
}

// NewBrokerHousekeepingService 初始化Broker活动检测服务
// Author: tianyuliang
// Since: 2017/9/6
func NewBrokerHousekeepingService(controller *DefaultNamesrvController) netm.ContextListener {
	brokerHousekeepingService := new(BrokerHousekeepingService)
	brokerHousekeepingService.NamesrvController = controller
	return brokerHousekeepingService
}

// OnContextConnect 创建Channel连接
// Author: tianyuliang
// Since: 2017/9/6
func (self *BrokerHousekeepingService) OnContextConnect(ctx netm.Context) {
	logger.Info("BrokerHousekeepingService.OnContextConnect() handle request. %s", ctx.ToString())
}

// OnContextClose 关闭Channel,通知Topic路由管理器，清除无效Broker
// Author: tianyuliang
// Since: 2017/9/6
func (self *BrokerHousekeepingService) OnContextClose(ctx netm.Context) {
	if ctx == nil {
		logger.Error("OnContextClose() net.conn is nil")
		return
	}
	logger.Warn("BrokerHousekeepingService.OnContextClose() handle request. %s", ctx.ToString())
	self.NamesrvController.RouteInfoManager.onChannelDestroy(ctx.RemoteAddr().String(), ctx)
}

// OnContextError Channel出现异常,通知Topic路由管理器，清除无效Broker
// Author: tianyuliang
// Since: 2017/9/6
func (self *BrokerHousekeepingService) OnContextError(ctx netm.Context) {
	if ctx == nil {
		logger.Error("OnContextError() net.conn is nil")
		return
	}
	logger.Warn("BrokerHousekeepingService.OnContextError() handle request. %s", ctx.ToString())
	self.NamesrvController.RouteInfoManager.onChannelDestroy(ctx.RemoteAddr().String(), ctx)
}

// OnContextIdle Channe的Idle时间超时,通知Topic路由管理器，清除无效Brokers
// Author: tianyuliang
// Since: 2017/9/6
func (self *BrokerHousekeepingService) OnContextIdle(ctx netm.Context) {
	if ctx == nil {
		logger.Error("OnContextIdle() net.conn is nil")
		return
	}
	logger.Warn("BrokerHousekeepingService.OnContextIdle() handle request. %s", ctx.ToString())
	self.NamesrvController.RouteInfoManager.onChannelDestroy(ctx.RemoteAddr().String(), ctx)
}
