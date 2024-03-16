package topic_tree

type SubscribeMode uint8

const (
	CommonlySubscribe   SubscribeMode = 0 // 普通的,没有通配符的订阅
	MonolayerSubscribe  SubscribeMode = 1 // 单层通配符—-“+”
	MultilayerSubscribe SubscribeMode = 2 // 多层通配符—-“#”
)

func newTopic(parent *topic, target string, ids ...int64) *topic {
	t := &topic{
		parent:    parent,
		target:    target,
		children:  map[string]*topic{},
		commonIds: map[int64]struct{}{},
		monoIds:   map[int64]struct{}{},
		multiIds:  map[int64]struct{}{},
	}
	if ids != nil && len(ids) > 0 {
		for _, id := range ids {
			t.commonIds[id] = struct{}{}
		}
	}
	return t
}

type topic struct {
	parent    *topic             // 上一级topic
	target    string             // 标识
	children  map[string]*topic  // 子topic
	commonIds map[int64]struct{} // 标准订阅
	monoIds   map[int64]struct{} // 单层通(+)订阅
	multiIds  map[int64]struct{} // 多层通(#)订阅
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
func (t *topic) Subscribe(clientId int64, mod SubscribeMode) {
	//t.mu.Lock()
	//defer t.mu.Unlock()
	switch mod {
	case CommonlySubscribe:
		if t.commonIds == nil {
			t.commonIds = map[int64]struct{}{}
		}
		t.commonIds[clientId] = struct{}{}
	case MonolayerSubscribe:
		if t.monoIds == nil {
			t.monoIds = map[int64]struct{}{}
		}
		t.monoIds[clientId] = struct{}{}
	case MultilayerSubscribe:
		if t.multiIds == nil {
			t.multiIds = map[int64]struct{}{}
		}
		t.multiIds[clientId] = struct{}{}
	}
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
			t.monoIds = map[int64]struct{}{}
		}
		m = t.monoIds
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
