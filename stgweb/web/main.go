package main

import (
	"flag"
	"fmt"
	"git.oschina.net/cloudzone/cloudcommon-go/web"
	"git.oschina.net/cloudzone/smartgo/stgweb/web/g"
	"git.oschina.net/cloudzone/smartgo/stgweb/web/route"
	"os"
	//"git.oschina.net/cloudzone/smartgo/stgcommon"
)

const (
	_version = "v1.0.0"
)

func main() {

	//os.Setenv(stgcommon.NAMESRV_ADDR_ENV, "10.112.68.190:9876;10.112.68.192:9876")
	//os.Setenv(stgcommon.BLOTMQ_WEB_CONFIG_ENV, "E:/source/src/git.oschina.net/cloudzone/smartgo/stgweb/web/etc/cfg.json")

	v := flag.Bool("v", false, "version")
	help := flag.Bool("h", false, "help")
	flag.Parse()

	if *v {
		fmt.Println(_version)
		os.Exit(0)
	}

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	g.Init()
	web.New(_version).Config(&g.GetConfig().Web).Call(func(ctx *web.Context) {
		ctx.Super().Action = route.Route
	}).End().Run()
}
