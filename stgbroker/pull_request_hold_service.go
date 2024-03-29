package stgbroker

import (
	"bytes"
	"fmt"
	"git.oschina.net/cloudzone/smartgo/stgbroker/longpolling"
	"git.oschina.net/cloudzone/smartgo/stgcommon/logger"
	"git.oschina.net/cloudzone/smartgo/stgcommon/sync"
	"git.oschina.net/cloudzone/smartgo/stgcommon/utils"
	"git.oschina.net/cloudzone/smartgo/stgcommon/utils/timeutil"
	"strconv"
	"strings"
	"time"
)

// PullRequestHoldService 拉消息请求管理，如果拉不到消息，则在这里Hold住，等待消息到来
// Author rongzhihong
// Since 2017/9/5
type PullRequestHoldService struct {
	TOPIC_QUEUEID_SEPARATOR string
	pullRequestTable        *sync.Map // key:topic@queueid value:ManyPullRequest
	brokerController        *BrokerController
	isStopped               bool
}

// NewPullRequestHoldService 初始化拉消息请求服务
// Author rongzhihong
// Since 2017/9/5
func NewPullRequestHoldService(brokerController *BrokerController) *PullRequestHoldService {
	holdServ := new(PullRequestHoldService)
	holdServ.pullRequestTable = sync.NewMap()
	holdServ.TOPIC_QUEUEID_SEPARATOR = TOPIC_GROUP_SEPARATOR
	holdServ.brokerController = brokerController
	return holdServ
}

// buildKey 构造Key
// Author rongzhihong
// Since 2017/9/5
func (serv *PullRequestHoldService) buildKey(topic string, queueId int32) string {
	sb := bytes.Buffer{}
	sb.WriteString(topic)
	sb.WriteString(serv.TOPIC_QUEUEID_SEPARATOR)
	sb.WriteString(fmt.Sprintf("%d", queueId))
	return sb.String()
}

// SuspendPullRequest 延缓拉请求
// Author rongzhihong
// Since 2017/9/5
func (serv *PullRequestHoldService) SuspendPullRequest(topic string, queueId int32, pullRequest *longpolling.PullRequest) {
	key := serv.buildKey(topic, queueId)
	mpr, err := serv.pullRequestTable.Get(key)
	if err != nil {
		logger.Error(err)
		return
	}

	if nil == mpr {
		mpr = new(longpolling.ManyPullRequest)
		prev, _ := serv.pullRequestTable.PutIfAbsent(key, mpr)
		if prev != nil {
			mpr = prev
		}
	}

	if bean, ok := mpr.(*longpolling.ManyPullRequest); ok {
		bean.AddPullRequest(pullRequest)
	}
}

// checkHoldRequest  检查拉请求是否有数据，如有，则通知
// Author rongzhihong
// Since 2017/9/5
func (serv *PullRequestHoldService) checkHoldRequest() {
	for iter := serv.pullRequestTable.Iterator(); iter.HasNext(); {
		key, _, _ := iter.Next()
		if item, ok := key.(string); ok {

			kArray := strings.Split(item, serv.TOPIC_QUEUEID_SEPARATOR)
			if kArray != nil && 2 == len(kArray) {
				topic := kArray[0]
				queueId, err := strconv.Atoi(kArray[1])
				if err != nil {
					logger.Errorf("queueId=%s: string to int fail.", kArray[1])
					continue
				}
				offset := serv.brokerController.MessageStore.GetMaxOffsetInQueue(topic, int32(queueId))
				serv.notifyMessageArriving(topic, int32(queueId), offset)
			}
		}
	}
}

// notifyMessageArriving  消息到来通知
// Author rongzhihong
// Since 2017/9/5
func (serv *PullRequestHoldService) notifyMessageArriving(topic string, queueId int32, maxOffset int64) {
	key := serv.buildKey(topic, queueId)
	mpr, err := serv.pullRequestTable.Get(key)
	if err != nil {
		logger.Error(err)
		return
	}
	if mpr == nil {
		return
	}

	if mpr, ok := mpr.(*longpolling.ManyPullRequest); ok {
		requestList := mpr.CloneListAndClear()
		if requestList == nil || len(requestList) <= 0 {
			return
		}

		replayList := []*longpolling.PullRequest{}
		for _, pullRequest := range requestList {
			// 查看是否offset OK
			if maxOffset > pullRequest.PullFromThisOffset {
				serv.brokerController.PullMessageProcessor.ExecuteRequestWhenWakeup(pullRequest.Context, pullRequest.RequestCommand)
				continue
			} else {
				// 尝试取最新Offset
				newestOffset := serv.brokerController.MessageStore.GetMaxOffsetInQueue(topic, queueId)
				if newestOffset > pullRequest.PullFromThisOffset {
					serv.brokerController.PullMessageProcessor.ExecuteRequestWhenWakeup(pullRequest.Context, pullRequest.RequestCommand)
					continue
				}
			}

			currentTimeMillis := timeutil.CurrentTimeMillis()
			// 查看是否超时
			if currentTimeMillis >= (pullRequest.SuspendTimestamp + pullRequest.TimeoutMillis) {
				serv.brokerController.PullMessageProcessor.ExecuteRequestWhenWakeup(pullRequest.Context, pullRequest.RequestCommand)
				continue
			}

			// 当前不满足要求，重新放回Hold列表中
			replayList = append(replayList, pullRequest)
		}

		if len(replayList) > 0 {
			mpr.AddManyPullRequest(replayList)
		}
	}
}

// run  运行入口
// Author rongzhihong
// Since 2017/9/5
func (serv *PullRequestHoldService) run() {
	defer utils.RecoveredFn()
	logger.Info(fmt.Sprintf("%s service started", serv.getServiceName()))

	for !serv.isStopped {
		time.Sleep(time.Millisecond * time.Duration(1000))
		serv.checkHoldRequest()
	}

	logger.Info(fmt.Sprintf("%s service end", serv.getServiceName()))
}

// Start  启动入口
// Author rongzhihong
// Since 2017/9/5
func (serv *PullRequestHoldService) Start() {
	go func() {
		serv.run()
	}()
	logger.Info("PullRequestHoldService start successful")
}

// Shutdown  停止
// Author rongzhihong
// Since 2017/9/5
func (serv *PullRequestHoldService) Shutdown() {
	serv.isStopped = true
	logger.Info("PullRequestHoldService shutdown successful")
}

// getServiceName  获得类名
// Author rongzhihong
// Since 2017/9/5
func (serv *PullRequestHoldService) getServiceName() string {
	return "PullRequestHoldService"
}
