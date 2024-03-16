package topic_tree

import "fmt"

type SubscribeMode uint8

const (
	CommonlySubscribe   SubscribeMode = 0 // 普通的,没有通配符的订阅
	MonolayerSubscribe  SubscribeMode = 1 // 单层通配符—-“+”
	MultilayerSubscribe SubscribeMode = 2 // 多层通配符—-“#”
)

func newTopic(parent *topic, title string, target string, ids ...int64) *topic {
	t := &topic{
		parent:    parent,
		title:     title,
		targets:   nil,
		level:     0,
		children:  map[string]*topic{},
		commonIds: map[int64]struct{}{},
		monoIds:   map[string]map[int64]struct{}{},
		multiIds:  map[string]map[int64]struct{}{},
	}
	if parent != nil {
		t.level = parent.level + 1
		if parent.multiIds != nil {
			if idMap, ok := parent.multiIds["#"]; ok {
				t.monoIds["#"] = idMap
			}
			key := fmt.Sprintf("#/%s", target)
			if idMap, ok := parent.multiIds[key]; ok {
				t.monoIds[key] = idMap
			}
		}
	}
	if ids != nil && len(ids) > 0 {
		for _, id := range ids {
			t.commonIds[id] = struct{}{}
		}
	}
	return t
}

type topic struct {
	parent    *topic                        // 上一级topic
	title     string                        // 标题
	targets   []string                      // 标识
	level     uint8                         // 级别
	children  map[string]*topic             // 子topic
	commonIds map[int64]struct{}            // 标准订阅
	monoIds   map[string]map[int64]struct{} // 单层通(+)订阅
	multiIds  map[string]map[int64]struct{} // 多层通(#)订阅
}

func (t *topic) UnSubscribe(clientId int64, mod SubscribeMode) {
	//t.mu.Lock()
	//defer t.mu.Unlock()
	switch mod {
	case CommonlySubscribe:
		if t.commonIds != nil {
			delete(t.commonIds, clientId)
		}

	case MonolayerSubscribe:
		if t.monoIds != nil {
			delete(t.monoIds, clientId)
		}
	case MultilayerSubscribe:
		if t.multiIds != nil {
			delete(t.multiIds, clientId)
		}
	}
}
func (t *topic) SubscribeMultilayer(clientId int64, multiStr string) {
	if t.multiIds == nil {
		t.multiIds = map[string]map[int64]struct{}{}
	}
	if _, ok := t.multiIds[multiStr]; !ok {
		t.multiIds[multiStr] = map[int64]struct{}{}
	}
	t.multiIds[multiStr][clientId] = struct{}{}
}
func (t *topic) SubscribeMonolayer(clientId int64, monoStr string) {
	if t.monoIds == nil {
		t.monoIds = map[int64]string{}
	}
	t.monoIds[clientId] = monoStr
}
func (t *topic) SubscribeCommonly(clientId int64) {
	//t.mu.Lock()
	//defer t.mu.Unlock()
	if t.commonIds == nil {
		t.commonIds = map[int64]struct{}{}
	}
	t.commonIds[clientId] = struct{}{}
	return

}

func (t *topic) GetIdsWithMode(mod SubscribeMode) map[int64]struct{} {
	//t.mu.Lock()
	//defer t.mu.Unlock()
	var m map[int64]struct{}
	switch mod {
	case CommonlySubscribe:
		if t.commonIds == nil {
			t.commonIds = map[int64]struct{}{}
		}
		m = t.commonIds
	case MonolayerSubscribe:
		if t.monoIds == nil {
			t.monoIds = map[int64]string{}
		} else {

		}

	case MultilayerSubscribe:
		if t.multiIds == nil {
			t.multiIds = map[int64]struct{}{}
		}
		m = t.multiIds
	}
	return m
}

func (t *topic) GetTopicMessageClientIds() []int64 {
	idMap := t.GetIdsWithMode(CommonlySubscribe)
	for id, _ := range t.GetIdsWithMode(MonolayerSubscribe) {
		idMap[id] = struct{}{}
	}
	for id, _ := range t.GetIdsWithMode(MonolayerSubscribe) {
		idMap[id] = struct{}{}
	}
	if t.parent != nil {
		for id, _ := range getParentIds(t.parent, true) {
			idMap[id] = struct{}{}
		}
	}
	var list []int64
	for id, _ := range idMap {
		list = append(list, id)
	}
	return list
}

func getParentIds(t *topic, isMono bool) map[int64]struct{} {
	m := t.GetIdsWithMode(MultilayerSubscribe)
	if isMono {
		for id, _ := range t.GetIdsWithMode(MonolayerSubscribe) {
			m[id] = struct{}{}
		}
	}
	if t.parent != nil {
		for id, _ := range getParentIds(t.parent, false) {
			m[id] = struct{}{}
		}
	}
	return m
}
