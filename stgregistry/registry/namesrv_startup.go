package registry

import (
	"fmt"
	"git.oschina.net/cloudzone/smartgo/stgcommon"
	"git.oschina.net/cloudzone/smartgo/stgcommon/namesrv"
	"git.oschina.net/cloudzone/smartgo/stgcommon/static"
	"git.oschina.net/cloudzone/smartgo/stgnet/remoting"
	"git.oschina.net/cloudzone/smartgo/stgregistry/logger"
	"os"
	"strconv"
)

// Startup 启动Namesrv控制器
// Author: tianyuliang
// Since: 2017/9/14
func Startup(stopChannel chan bool, registryPort int) *DefaultNamesrvController {
	// 构建NamesrvController
	controller := CreateNamesrvController(registryPort)

	// NamesrvController初始化
	initResult := controller.initialize()
	if !initResult {
		fmt.Println("the name server controller initialize failed")
		controller.shutdown()
		os.Exit(0)
	}

	// 注册ShutdownHook钩子
	controller.registerShutdownHook(stopChannel)

	// 启动
	go func() {
		// 额外处理“RemotingServer.Stacr()启动后，导致channel缓冲区满，进而引发namesrv主线程阻塞”情况
		controller.startNamesrvController()
	}()
	fmt.Println("the name server boot success") // 此处不要使用logger.Info(),给nohup.out提示

	return controller
}

// CreateNamesrvController 创建默认Namesrv控制器
// Author: tianyuliang
// Since: 2017/9/15
func CreateNamesrvController(registryPort int) *DefaultNamesrvController {
	cfg := namesrv.NewNamesrvConfig()
	logger.Info("%s", cfg.ToString())

	listenPort := static.REGISTRY_PORT
	if registryPort > 0 {
		listenPort = registryPort
	}
	if namesrvPort, err := strconv.Atoi(stgcommon.GetNamesrvPort()); err == nil && namesrvPort > 0 {
		listenPort = namesrvPort
	}
	remotingServer := remoting.NewDefalutRemotingServer(static.REGISTRY_IP, listenPort)
	controller := NewNamesrvController(cfg, remotingServer)

	logger.Info("create name server controller success. listenPort=%d", listenPort)
	return controller
}
