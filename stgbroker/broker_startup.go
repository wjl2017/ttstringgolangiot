package stgbroker

import (
	"fmt"
	"git.oschina.net/cloudzone/smartgo/stgcommon"
	"git.oschina.net/cloudzone/smartgo/stgcommon/logger"
	"git.oschina.net/cloudzone/smartgo/stgcommon/static"
	"git.oschina.net/cloudzone/smartgo/stgcommon/utils"
	"git.oschina.net/cloudzone/smartgo/stgcommon/utils/parseutil"
	"git.oschina.net/cloudzone/smartgo/stgnet/remoting"
	"git.oschina.net/cloudzone/smartgo/stgstorelog"
	"git.oschina.net/cloudzone/smartgo/stgstorelog/config"
	"github.com/toolkits/file"
	"os"
	"path/filepath"
	"strings"
)

// Start 启动BrokerController
// Author: tianyuliang
// Since: 2017/9/20
func Start(stopChan chan bool, smartgoBrokerFilePath string) *BrokerController {
	// 构建BrokerController控制器、初始化BrokerController
	controller := CreateBrokerController(smartgoBrokerFilePath)

	// 注册ShutdownHook钩子
	controller.registerShutdownHook(stopChan)

	// 初始化broker必要的资源
	controller.Initialize()

	// 启动BrokerController
	controller.Start()

	format := "the broker[%s, %s] boot success."
	tips := fmt.Sprintf(format, controller.BrokerConfig.BrokerName, controller.GetBrokerAddr())
	if controller.BrokerConfig.NamesrvAddr != "" {
		format = "the broker[%s, %s] boot success, and the name server is %s"
		tips = fmt.Sprintf(format, controller.BrokerConfig.BrokerName, controller.GetBrokerAddr(), controller.BrokerConfig.NamesrvAddr)
	}
	fmt.Println(tips) // 此处不要使用logger.Info(),给nohup.out提示

	return controller
}

// CreateBrokerController 创建BrokerController对象
// Author: tianyuliang
// Since: 2017/9/20
func CreateBrokerController(smartgoBrokerFilePath ...string) *BrokerController {
	defer utils.RecoveredFn()
	cfgName := getSmartgoBrokerConfigName(smartgoBrokerFilePath...)
	cfgPath := getSmartgoBrokerConfigPath(cfgName)

	// 读取并转化*.toml配置项的值
	cfg, ok := parseSmartgoBrokerConfig(cfgName, cfgPath)
	if !ok || cfg == nil {
		logger.Flush()
		os.Exit(0)
	}

	// 初始化brokerConfig，并校验broker启动的所必需的SmartGoHome、Namesrv配置
	brokerConfig := stgcommon.NewCustomBrokerConfig(cfg)
	logger.Infof("broker.StorePathRootDir = %s", brokerConfig.StorePathRootDir)
	logger.Infof("store.StorePathRootDir = %s", brokerConfig.StorePathRootDir)

	if !brokerConfig.CheckBrokerConfigAttr(smartgoBrokerFilePath...) {
		logger.Flush()
		os.Exit(0)
	}

	brorkerRole, err := config.ParseBrokerRole(cfg.BrokerRole)
	if err != nil {
		logger.Errorf(err.Error())
		logger.Flush()
		os.Exit(0)
	}

	// 初始化brokerConfig、messageStoreConfig
	messageStoreConfig := stgstorelog.NewMessageStoreConfig()
	messageStoreConfig.BrokerRole = brorkerRole
	if !checkMessageStoreConfigAttr(messageStoreConfig, brokerConfig) {
		logger.Flush()
		os.Exit(0)
	}
	err = setMessageStoreConfig(messageStoreConfig, brokerConfig, cfg)
	if err != nil {
		logger.Errorf(err.Error())
		logger.Flush()
		os.Exit(0)
	}

	// 构建BrokerController结构体
	remotingClient := remoting.NewDefalutRemotingClient()
	controller := NewBrokerController(brokerConfig, messageStoreConfig, remotingClient)
	controller.ConfigFile = brokerConfig.StorePathRootDir

	logger.Info("create broker controller successful")
	return controller
}

// parseSmartgoBrokerConfig 读取并转化Broker启动所必须的配置文件
//
// 注意：broker、store等模块的存取数据目录(优先级从高到低)
// (1)smartgoBroker.toml.SmartgoDataPath
// (2)$SMARTGO_DATA_PATH
// (3)user.Current().HomeDir
//
// Author: tianyuliang
// Since: 2017/9/22
func parseSmartgoBrokerConfig(cfgName, cfgPath string) (*stgcommon.SmartgoBrokerConfig, bool) {
	// 读取并转化*.toml配置项的值
	var cfg stgcommon.SmartgoBrokerConfig
	parseutil.ParseConf(cfgPath, &cfg)
	if &cfg == nil {
		logger.Errorf("read %s error", cfgPath)
		return nil, false
	}

	logger.Info(cfg.ToString())
	if cfg.IsBlank() {
		logger.Errorf("read broker toml failed. %s", cfgPath)
		return nil, false
	}

	if strings.TrimSpace(cfg.StorePathRootDir) == "" {
		cfg.StorePathRootDir = stgcommon.GetUserHomeDir() + separator + static.BROKER_DATA_ROOT_DIR
	}
	return &cfg, true
}

// getSmartgoBrokerConfigName 获得启动broker的toml文件名称
// Author: tianyuliang
// Since: 2017/10/16
func getSmartgoBrokerConfigName(smartgoBrokerFilePath ...string) string {
	defer utils.RecoveredFn()

	if smartgoBrokerFilePath != nil && len(smartgoBrokerFilePath) > 0 && smartgoBrokerFilePath[0] != "" {
		value := filepath.ToSlash(strings.TrimSpace(smartgoBrokerFilePath[0]))
		index := strings.LastIndex(value, "/")
		cfgName := value[index+1 : len(value)]
		return cfgName
	}
	return static.BROKER_CONFIG_NAME
}

// getSmartgoBrokerConfigPath 获得启动broker的toml文件完整路径
//
// 注意：toml文件完整路径优先级(如下从高到低): 带有$SMARTGO_HOME的路径优先级最高、最末尾smartgo源码路径的优先级最低
// (1)$SMARTGO_HOME/conf/broker-a.toml
// (2)./conf/broker-a.toml
// (3)../../conf/broker-a.toml
// (4)$GOPATH/src/git.oschina.net/cloudzone/smartgo/conf/broker-a.toml
//
// Author: tianyuliang
// Since: 2017/10/16
func getSmartgoBrokerConfigPath(cfgName string) string {
	cfgPath := stgcommon.GetSmartGoHome() + "/conf/" + cfgName // 各种main()启动broker,读取环境变量对应的路径
	if file.IsExist(cfgPath) {
		logger.Infof("environment brokerConfigPath = %s", cfgPath)
		return cfgPath
	}

	cfgPath = file.SelfDir() + "/conf/" + cfgName // 各种部署目录
	if file.IsExist(cfgPath) {
		logger.Infof("current brokerConfigPath = %s", cfgPath)
		return cfgPath
	}

	cfgPath = "../../conf/" + cfgName // 各种test用例启动broker,读取相对路径
	if file.IsExist(cfgPath) {
		logger.Infof("test case brokerConfigPath = %s", cfgPath)
		return cfgPath
	}

	// 在IDEA上面利用conf/smartgoBroker.toml默认配置文件目录
	cfgPath = stgcommon.GetSmartgoConfigDir() + cfgName
	logger.Infof("idea special brokerConfigPath = %s", cfgPath)
	return cfgPath
}

// checkMessageStoreConfigAttr 校验messageStoreConfig配置
// Author: tianyuliang
// Since: 2017/9/22
func checkMessageStoreConfigAttr(mscfg *stgstorelog.MessageStoreConfig, bcfg *stgcommon.BrokerConfig) bool {
	if mscfg.BrokerRole == config.SLAVE {
		if bcfg.BrokerId <= 0 {
			logger.Errorf("Slave's brokerId[%d] must be > 0", bcfg.BrokerId)
			return false
		}
		if bcfg.HaMasterAddress == "" || !stgcommon.CheckIpAndPort(bcfg.HaMasterAddress) {
			logger.Errorf("Slave's HaMasterAddress[%s] invalid", bcfg.HaMasterAddress)
			return false
		}
	}
	return true
}

// setMessageStoreConfig 设置messageStoreConfig配置
// Author: tianyuliang
// Since: 2017/9/22
func setMessageStoreConfig(messageStoreConfig *stgstorelog.MessageStoreConfig, brokerConfig *stgcommon.BrokerConfig, cfg *stgcommon.SmartgoBrokerConfig) error {
	// 此处需要覆盖store模块的StorePathRootDir配置目录,用来处理一台服务器启动多个broker的场景
	messageStoreConfig.StorePathRootDir = brokerConfig.StorePathRootDir
	messageStoreConfig.StorePathCommitLog = brokerConfig.StorePathRootDir + separator + static.STORE_COMMIT_LOG_ROOT_DIR
	if brokerConfig.HaMasterAddress != "" {
		messageStoreConfig.HaMasterAddress = brokerConfig.HaMasterAddress // HA功能配置此项

	}

	// 如果是slave，修改默认值（修改命中消息在内存的最大比例40为30【40-10】）
	if messageStoreConfig.BrokerRole == config.SLAVE {
		ratio := messageStoreConfig.AccessMessageInMemoryMaxRatio - 10
		messageStoreConfig.AccessMessageInMemoryMaxRatio = ratio
	}

	flushDiskType, err := config.ParseFlushDiskType(cfg.FlushDiskType)
	if err != nil {
		return err
	}
	messageStoreConfig.FlushDiskType = flushDiskType

	// BrokerId的处理 switch-case语法：
	// 只要匹配到一个case，则顺序往下执行，直到遇到break，因此若没有break则不管后续case匹配与否都会执行
	switch messageStoreConfig.BrokerRole {
	//如果是同步master也会执行下述case中brokerConfig.setBrokerId(MixAll.MASTER_ID);语句，直到遇到break
	case config.ASYNC_MASTER:
		fallthrough
	case config.SYNC_MASTER:
		brokerConfig.BrokerId = stgcommon.MASTER_ID
	case config.SLAVE:
	default:

	}

	return nil
}
