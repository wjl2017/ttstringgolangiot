package main

import (
	"fmt"
	code "git.oschina.net/cloudzone/smartgo/stgcommon/protocol"
	"git.oschina.net/cloudzone/smartgo/stgcommon/protocol/header/namesrv"
	"git.oschina.net/cloudzone/smartgo/stgnet/netm"
	"git.oschina.net/cloudzone/smartgo/stgnet/protocol"
	"git.oschina.net/cloudzone/smartgo/stgnet/remoting"
	"log"
)

var (
	remotingServer remoting.RemotingServer
)

type GetTopicStatsInfoProcessor struct {
}

func (processor *GetTopicStatsInfoProcessor) ProcessRequest(ctx netm.Context,
	request *protocol.RemotingCommand) (*protocol.RemotingCommand, error) {
	fmt.Printf("GetTopicStatsInfoProcessor %d %d\n", request.Code, request.Opaque)

	topicStatsInfoRequestHeader := &namesrv.GetTopicStatsInfoRequestHeader{}
	err := request.DecodeCommandCustomHeader(topicStatsInfoRequestHeader)
	if err != nil {
		return nil, err
	}
	fmt.Printf("DecodeCommandCustomHeader %v\n", topicStatsInfoRequestHeader)

	response := protocol.CreateResponseCommand(code.SUCCESS, "success")
	response.Opaque = request.Opaque

	return response, nil
}

type OtherProcessor struct {
}

func (processor *OtherProcessor) ProcessRequest(ctx netm.Context,
	request *protocol.RemotingCommand) (*protocol.RemotingCommand, error) {
	fmt.Printf("OtherProcessor %d %d\n", request.Code, request.Opaque)

	response := protocol.CreateResponseCommand(code.SUCCESS, "success")
	response.Opaque = request.Opaque

	return response, nil
}

type ServerContextListener struct {
}

func (listener *ServerContextListener) OnContextConnect(ctx netm.Context) {
	log.Printf("one connection create: addr[%s] localAddr[%s] remoteAddr[%s]\n", ctx.Addr(), ctx.LocalAddr(), ctx.RemoteAddr())
}

func (listener *ServerContextListener) OnContextClose(ctx netm.Context) {
	log.Printf("one connection close: addr[%s] localAddr[%s] remoteAddr[%s]\n", ctx.Addr(), ctx.LocalAddr(), ctx.RemoteAddr())
}

func (listener *ServerContextListener) OnContextError(ctx netm.Context) {
	log.Printf("one connection error: addr[%s] localAddr[%s] remoteAddr[%s]\n", ctx.Addr(), ctx.LocalAddr(), ctx.RemoteAddr())
}

func (listener *ServerContextListener) OnContextIdle(ctx netm.Context) {
	log.Printf("one connection idle: addr[%s] localAddr[%s] remoteAddr[%s]\n", ctx.Addr(), ctx.LocalAddr(), ctx.RemoteAddr())
}

func main() {
	initServer()
	remotingServer.Start()
}

func initServer() {
	remotingServer = remoting.NewDefalutRemotingServer("0.0.0.0", 10911)
	remotingServer.RegisterProcessor(code.HEART_BEAT, &OtherProcessor{})
	remotingServer.RegisterProcessor(code.SEND_MESSAGE_V2, &OtherProcessor{})
	remotingServer.RegisterProcessor(code.GET_TOPIC_STATS_INFO, &GetTopicStatsInfoProcessor{})
	remotingServer.RegisterProcessor(code.GET_MAX_OFFSET, &OtherProcessor{})
	remotingServer.RegisterProcessor(code.QUERY_CONSUMER_OFFSET, &OtherProcessor{})
	remotingServer.RegisterProcessor(code.PULL_MESSAGE, &OtherProcessor{})
	remotingServer.RegisterProcessor(code.UPDATE_CONSUMER_OFFSET, &OtherProcessor{})
	remotingServer.RegisterProcessor(code.GET_CONSUMER_LIST_BY_GROUP, &OtherProcessor{})
	remotingServer.RegisterProcessor(code.GET_ROUTEINTO_BY_TOPIC, &OtherProcessor{})
	remotingServer.RegisterProcessor(code.UPDATE_AND_CREATE_TOPIC, &OtherProcessor{})
	remotingServer.RegisterProcessor(code.GET_KV_CONFIG, &OtherProcessor{})
	remotingServer.RegisterContextListener(&ServerContextListener{})
}
