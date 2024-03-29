package remoting

import (
	"strconv"

	"git.oschina.net/cloudzone/smartgo/stgnet/netm"
	"git.oschina.net/cloudzone/smartgo/stgnet/protocol"
)

// DefalutRemotingServer default remoting server
type DefalutRemotingServer struct {
	host      string
	port      int
	bootstrap *netm.Bootstrap
	BaseRemotingAchieve
}

// NewDefalutRemotingServer return new default remoting server
func NewDefalutRemotingServer(host string, port int) *DefalutRemotingServer {
	remotingServe := &DefalutRemotingServer{
		host: host,
		port: port,
	}
	remotingServe.responseTable = make(map[int32]*ResponseFuture)
	remotingServe.fragmentationActuator = NewLengthFieldFragmentationAssemblage(FRAME_MAX_LENGTH, 0, 4, 0)
	remotingServe.bootstrap = netm.NewBootstrap()
	return remotingServe
}

// Start start server
func (rs *DefalutRemotingServer) Start() {
	rs.bootstrap.RegisterHandler(func(buffer []byte, ctx netm.Context) {
		rs.processReceived(buffer, ctx)
	})

	rs.isRunning = true
	// 定时扫描响应
	rs.startScheduledTask()

	// 启动服务
	rs.bootstrap.Bind(rs.host, rs.port).Sync()
}

// Shutdown shutdown server
func (rs *DefalutRemotingServer) Shutdown() {
	if rs.timeoutTimer != nil {
		rs.timeoutTimer.Stop()
	}

	if rs.bootstrap != nil {
		rs.bootstrap.Shutdown()
	}
	rs.isRunning = false
}

// InvokeSync 同步调用并返回响应, addr为空字符串
func (rs *DefalutRemotingServer) InvokeSync(ctx netm.Context, request *protocol.RemotingCommand, timeoutMillis int64) (*protocol.RemotingCommand, error) {
	//addr := ctx.RemoteAddr().String()
	return rs.invokeSync(ctx, request, timeoutMillis)
}

// InvokeAsync 异步调用
func (rs *DefalutRemotingServer) InvokeAsync(ctx netm.Context, request *protocol.RemotingCommand, timeoutMillis int64, invokeCallback InvokeCallback) error {
	//addr := ctx.RemoteAddr().String()
	return rs.invokeAsync(ctx, request, timeoutMillis, invokeCallback)
}

// InvokeSync 单向发送消息
func (rs *DefalutRemotingServer) InvokeOneway(ctx netm.Context, request *protocol.RemotingCommand, timeoutMillis int64) error {
	//addr := ctx.RemoteAddr().String()
	return rs.invokeOneway(ctx, request, timeoutMillis)
}

// GetListenPort 获得监听端口字符串
// Author rongzhihong
// Since 2017/9/5
func (rs *DefalutRemotingServer) GetListenPort() string {
	return strconv.Itoa(rs.port)
}

// Port 获得监听端口
// Author rongzhihong
// Since 2017/9/5
func (rs *DefalutRemotingServer) Port() int32 {
	return int32(rs.port)
}

// RegisterContextListener 注册context listener
func (rs *DefalutRemotingServer) RegisterContextListener(contextListener netm.ContextListener) {
	rs.bootstrap.RegisterContextListener(contextListener)
}
