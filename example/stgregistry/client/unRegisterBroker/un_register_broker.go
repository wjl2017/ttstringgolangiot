package main

import (
	"fmt"
	"git.oschina.net/cloudzone/smartgo/example/stgregistry/client"
	code "git.oschina.net/cloudzone/smartgo/stgcommon/protocol"
	"git.oschina.net/cloudzone/smartgo/stgcommon/protocol/header/namesrv"
	"git.oschina.net/cloudzone/smartgo/stgnet/protocol"
	"git.oschina.net/cloudzone/smartgo/stgnet/remoting"
	"git.oschina.net/cloudzone/smartgo/stgregistry/logger"
)

var (
	cmd remoting.RemotingClient
)

func initClient() {
	cmd = remoting.NewDefalutRemotingClient()
	cmd.UpdateNameServerAddressList([]string{client.DEFAULT_NAMESRV})
}

func main() {
	var (
		request     *protocol.RemotingCommand
		response    *protocol.RemotingCommand
		err         error
		brokerName  = "broker-b"
		brokerAddr  = "10.122.2.28:10911"
		clusterName = "DefaultCluster"
		brokerId    = 0
	)

	// 初始化
	initClient()

	// 启动
	cmd.Start()
	fmt.Println("remoting client start success")

	// 请求的custom header
	requestHeader := namesrv.NewUnRegisterBrokerRequestHeader(brokerName, brokerAddr, clusterName, brokerId)
	request = protocol.CreateRequestCommand(code.UNREGISTER_BROKER, requestHeader)

	namesrvAddrs := cmd.GetNameServerAddressList()
	if namesrvAddrs != nil && len(namesrvAddrs) > 0 {
		for _, namesrvAddr := range namesrvAddrs {
			// 同步发送请求
			response, err = cmd.InvokeSync(namesrvAddr, request, client.DEFAULT_TIMEOUT)
			if err != nil {
				logger.Error("sync response UNREGISTER_BROKER failed. err: %s", err.Error())
				return
			}
			if response == nil {
				logger.Error("sync response UNREGISTER_BROKER failed. err: response is nil")
				return
			}
			if response.Code == code.SUCCESS {
				logger.Info("sync response UNREGISTER_BROKER success.")
				return
			}
			format := "sync handle UNREGISTER_BROKER failed. code=%d, remark=%s"
			logger.Info(format, response.Code, response.Remark)
		}
	}

}
