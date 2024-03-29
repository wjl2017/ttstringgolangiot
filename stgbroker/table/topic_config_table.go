package table

import (
	"git.oschina.net/cloudzone/smartgo/stgcommon"
	"sync"
)

type TopicConfigTable struct {
	TopicConfigs map[string]*stgcommon.TopicConfig `json:"TopicConfigTable"`
	sync.RWMutex `json:"-"`
}

func NewTopicConfigTable() *TopicConfigTable {
	return &TopicConfigTable{
		TopicConfigs: make(map[string]*stgcommon.TopicConfig),
	}
}

func (table *TopicConfigTable) Size() int {
	table.RLock()
	defer table.RUnlock()

	return len(table.TopicConfigs)
}

func (table *TopicConfigTable) Put(k string, v *stgcommon.TopicConfig)*stgcommon.TopicConfig {
	table.Lock()
	defer table.Unlock()
	table.TopicConfigs[k] = v
	return v
}

func (table *TopicConfigTable) Get(k string) *stgcommon.TopicConfig {
	table.RLock()
	defer table.RUnlock()

	v, ok := table.TopicConfigs[k]
	if !ok {
		return nil
	}

	return v
}

func (table *TopicConfigTable) Remove(k string) *stgcommon.TopicConfig {
	table.Lock()
	defer table.Unlock()

	v, ok := table.TopicConfigs[k]
	if !ok {
		return nil
	}

	delete(table.TopicConfigs, k)
	return v
}

func (table *TopicConfigTable) Foreach(fn func(k string, v *stgcommon.TopicConfig)) {
	table.RLock()
	defer table.RUnlock()

	for k, v := range table.TopicConfigs {
		fn(k, v)
	}
}
