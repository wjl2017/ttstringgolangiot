package stgbroker

import (
	syncmap "git.oschina.net/cloudzone/smartgo/stgcommon/sync"
	"sync"
)

type OffsetTable struct {
	Offsets      map[string]map[int]int64 `json:"offsets"`
	sync.RWMutex `json:"-"`
}

func newOffsetTable() *OffsetTable {
	return &OffsetTable{
		Offsets: make(map[string]map[int]int64),
	}
}

func (table *OffsetTable) Size() int {
	table.RLock()
	defer table.RUnlock()

	return len(table.Offsets)
}

func (table *OffsetTable) Put(k string, v map[int]int64) {
	table.Lock()
	defer table.Unlock()
	table.Offsets[k] = v
}

func (table *OffsetTable) Get(k string) map[int]int64 {
	table.RLock()
	defer table.RUnlock()

	v, ok := table.Offsets[k]
	if !ok {
		return nil
	}

	return v
}

func (table *OffsetTable) Remove(k string) map[int]int64 {
	table.Lock()
	defer table.Unlock()

	v, ok := table.Offsets[k]
	if !ok {
		return nil
	}

	delete(table.Offsets, k)
	return v
}

func (table *OffsetTable) Foreach(fn func(k string, v map[int]int64)) {
	table.RLock()
	defer table.RUnlock()

	for k, v := range table.Offsets {
		fn(k, v)
	}
}

func (table *OffsetTable) RemoveByFlag(fn func(k string, v map[int]int64) bool) {
	table.Lock()
	for k, v := range table.Offsets {
		if fn(k, v) {
			delete(table.Offsets, k)
		}
	}
	table.Unlock()
}

// PutAll 同步Offset配置文件
// Author rongzhihong
// Since 2017/9/18
func (table *OffsetTable) PutAll(offsetMap *syncmap.Map) {
	table.Lock()
	defer table.Unlock()

	for iter := offsetMap.Iterator(); iter.HasNext(); {
		kItem, vItem, _ := iter.Next()
		var k string = ""
		var ok bool = false
		if k, ok = kItem.(string); !ok {
			continue
		}
		if v, vok := vItem.(map[int]int64); vok {
			table.Offsets[k] = v
		}
	}
}
