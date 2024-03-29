package filter

import (
	"git.oschina.net/cloudzone/smartgo/stgcommon/protocol/heartbeat"
	set "github.com/deckarep/golang-set"
	"strings"
	"errors"
	"git.oschina.net/cloudzone/smartgo/stgcommon"
)
// FilterAPI: filter api
// Author: yintongqiang
// Since:  2017/8/11

type FilterAPI struct {

}

func BuildSubscriptionData(consumerGroup string, topic string, subString string) (*heartbeat.SubscriptionData, error) {
	subscriptionData := &heartbeat.SubscriptionData{Topic:topic, SubString:subString, TagsSet:set.NewSet(), CodeSet:set.NewSet()}
	if strings.EqualFold(subString, "") || strings.EqualFold(subString, "*") {
		subscriptionData.SubString = "*"
	} else {
		tags := strings.Split(subString, "||")
		for _, tag := range tags {
			trimTag := strings.TrimSpace(tag)
			if !strings.EqualFold(trimTag, "") {
				subscriptionData.TagsSet.Add(trimTag)
				subscriptionData.CodeSet.Add(stgcommon.HashCode(trimTag))
			} else {
				return subscriptionData, errors.New("subString split error")
			}
		}
	}
	return subscriptionData, nil
}

func BuildSubscriptionData4Ponit(consumerGroup string, topic string, subString string) (subscription *heartbeat.SubscriptionData, err error) {
	subscriptionData := &heartbeat.SubscriptionData{Topic: topic, SubString: subString, TagsSet: set.NewSet(), CodeSet: set.NewSet()}
	if strings.EqualFold(subString, "") || strings.EqualFold(subString, "*") {
		subscriptionData.SubString = "*"
	} else {
		tags := strings.Split(subString, "||")
		for _, tag := range tags {
			trimTag := strings.TrimSpace(tag)
			if !strings.EqualFold(trimTag, "") {
				subscriptionData.TagsSet.Add(trimTag)
				subscriptionData.CodeSet.Add(stgcommon.HashCode(trimTag))
			} else {
				return subscriptionData, errors.New("subString split error")
			}
		}
	}
	return subscriptionData, nil
}
