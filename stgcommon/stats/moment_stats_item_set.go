package stats

import (
	"git.oschina.net/cloudzone/smartgo/stgcommon"
	"git.oschina.net/cloudzone/smartgo/stgcommon/utils"
	"git.oschina.net/cloudzone/smartgo/stgcommon/utils/timeutil"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

// MomentStatsItemSet  QueueId@Topic@Group的offset落后数量的记录集合
// Author rongzhihong
// Since 2017/9/17
type MomentStatsItemSet struct {
	sync.RWMutex
	StatsName              string                      `json:"statsName"`
	StatsItemTable         map[string]*MomentStatsItem `json:"statsItemTable"`
	MomentStatsTaskTickers *timeutil.Tickers           // broker统计的定时任务
}

// NewMomentStatsItemSet 初始化
// Author rongzhihong
// Since 2017/9/17
func NewMomentStatsItemSet(statsName string) *MomentStatsItemSet {
	item := new(MomentStatsItemSet)
	item.StatsName = statsName
	item.StatsItemTable = make(map[string]*MomentStatsItem, 128)
	item.MomentStatsTaskTickers = timeutil.NewTickers()

	item.init()
	return item
}

// GetAndCreateStatsItem  GetAndCreateStatsItem
// Author rongzhihong
// Since 2017/9/19
func (moment *MomentStatsItemSet) GetAndCreateStatsItem(statsKey string) *MomentStatsItem {
	moment.Lock()
	defer moment.Unlock()
	defer utils.RecoveredFn()

	statsItem := moment.StatsItemTable[statsKey]
	if nil == statsItem {
		statsItem = NewMomentStatsItem()
		statsItem.StatsName = moment.StatsName
		statsItem.StatsKey = statsKey
		moment.StatsItemTable[statsKey] = statsItem
	}
	return statsItem
}

// SetValue  statsKey的数值加value
// Author rongzhihong
// Since 2017/9/19
func (moment *MomentStatsItemSet) SetValue(statsKey string, value int64) {
	statsItem := moment.GetAndCreateStatsItem(statsKey)
	atomic.AddInt64(&(statsItem.ValueCounter), value)
}

// init  init
// Author rongzhihong
// Since 2017/9/19
func (moment *MomentStatsItemSet) init() {
	diffMin := float64(stgcommon.ComputNextMinutesTimeMillis() - timeutil.CurrentTimeMillis())
	var delayMin int = int(math.Abs(diffMin))

	moment.MomentStatsTaskTickers.Register("momentStatsItemSet_printAtMinutesTicker",
		timeutil.NewTicker(false, time.Duration(delayMin)*time.Millisecond, 5*time.Minute, func() { moment.printAtMinutes() }))
}

// printAtMinutes  输出每分钟数据
// Author rongzhihong
// Since 2017/9/19
func (moment *MomentStatsItemSet) printAtMinutes() {
	moment.RLock()
	defer moment.RUnlock()
	defer utils.RecoveredFn()

	for _, momentStatsItem := range moment.StatsItemTable {
		momentStatsItem.printAtMinutes()
	}
}
