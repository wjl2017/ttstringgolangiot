package g

import (
	"fmt"
	"git.oschina.net/cloudzone/smartgo/stgcommon"
	"git.oschina.net/cloudzone/smartgo/stgcommon/utils"
	"git.oschina.net/cloudzone/smartgo/stgregistry/logger"
	"github.com/toolkits/file"
	"log"
	"os"
)

type Config struct {
	Log logger.Config `json:"log"`
}

const (
	cfgName             = "cfg.json"
	cacheSize           = 10000
	loggerFuncCallDepth = 3
	enableFuncCallDepth = false
	maxdays             = 30
	level               = 6
	filename            = "./logs/registry.log"
)

var (
	cfg Config
)

// Init 模块初始化
// Author: tianyuliang
// Since: 2017/9/21
func InitLogger(configPath string) {
	cfgPath := configPath
	if cfgPath == "" {
		cfgPath = getLoggerConfigPath()
	}

	err := utils.ParseConfig(cfgPath, &cfg)
	if err == nil {
		log.Printf("read config file %s success \n", cfgPath)
		logger.SetConfig(cfg.Log)
	} else {
		log.Printf("set default config to logger")
		logger.SetConfig(getDefaultLoggerConfig())
	}
}

// getLoggerConfigPath 获取日志配置文件路径
// Author: tianyuliang
// Since: 2017/9/21
func getLoggerConfigPath() (cfgPath string) {
	// export SMARTGO_REGISTRY_CONFIG = "/home/registry/cfg.json"
	cfgPath = stgcommon.GetSmartRegistryConfig()
	if file.IsExist(cfgPath) {
		return cfgPath
	}

	// 默认寻找当前目录的cfg.json日志配置文件
	cfgPath = file.SelfDir() + string(os.PathSeparator) + cfgName
	if file.IsExist(cfgPath) {
		return cfgPath
	}

	// 此处为了兼容能够直接在idea上面利用start/g/默认配置文件目录
	cfgPath = stgcommon.GetSmartgoConfigDir(cfg) + cfgName
	fmt.Printf("idea special registryConfigPath = %s \n", cfgPath)
	return cfgPath
}

// getDefaultLoggerConfig 获得默认logger配置
// Author: tianyuliang
// Since: 2017/9/21
func getDefaultLoggerConfig() logger.Config {
	config := logger.Config{}
	config.CacheSize = cacheSize
	config.EnableFuncCallDepth = enableFuncCallDepth
	config.FuncCallDepth = loggerFuncCallDepth

	param := make(map[string]interface{})
	//param["filename"] = filename                      // 保存日志文件地址
	param["level"] = level                             // 日志级别 6:info, 7:debug
	param["maxdays"] = maxdays                         // 每天一个日志文件，最多保留文件个数
	param["enableFuncCallDepth"] = enableFuncCallDepth // 打印日志，是否显示文件名
	param["loggerFuncCallDepth"] = loggerFuncCallDepth // 堆栈日志打印层数
	config.Engine.Config = param                       // 转化过去
	config.Engine.Adapter = loggerType.Console         // 日志类型 文件file、 控制台console

	return config
}

// loggerType 日志保存类型：文件、控制台
// Author: tianyuliang
// Since: 2017/9/21
var loggerType = struct {
	File    string // 保存文件
	Console string // 打印到控制台
}{
	"file",
	"console",
}
