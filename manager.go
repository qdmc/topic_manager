package topic_manager

import (
	"errors"
	"fmt"
	"sort"
	"sync"
)

type SubscribeHandle func(title string, id ClientId) error

type SubscribeResultItem struct {
	SubscribeTitle string
	TopicTitle     string
	Result         SubscribeLevel
	Err            error
}

type UnSubscribeResultItem struct {
	Title string
	Err   error
}

// TopicManagerInterface      主题管理器通用接口
type TopicManagerInterface interface {
	ClientSubscribe(id ClientId, titles []string, levels ...SubscribeLevel) []SubscribeResultItem // 客户端订阅
	ClientUnSubscribe(id ClientId, titles []string) ([]UnSubscribeResultItem, error)              // 客户端取消订阅
	ClientUnSubscribeAll(id ClientId)                                                             // 客户端取消所有订阅
	GetClientSubscribe(id ClientId, start, end int) (int, []SubscribeItem, error)                 // 客户端订阅列表
	GetPublishClientIds(title string, isCheckTopicExists ...bool) (map[ClientId]struct{}, error)  // 消息发布受众列表,isCheckTopicExists:是否校验topic是否存在
	GetPlainTopics(start, end int) (int, []TopicInterface)                                        // 普通topic列表
	GetMatchTopics(start, end int) (int, []TopicInterface)                                        // 匹配topic列表
	SetSubscribeHandle(SubscribeHandle)                                                           // 配置一个校验订阅的Handle
	GetOnceTopicSubscribes(title string, start, end int) (int, []SubscribeItem, error)            // 获取一个topic的订阅列表
}

type SubscribeItem struct {
	Id       ClientId
	Title    string
	TimeNano int64
}

type ClientItem struct {
	id       ClientId
	topicMap map[string]TopicInterface
}

func NewTopicManager() TopicManagerInterface {
	return &defaultTopicManager{
		mu:        sync.RWMutex{},
		clientMap: map[ClientId]ClientItem{},
		plainMap:  map[string]TopicInterface{},
		matchMap:  map[string]TopicInterface{},
		handle: func(title string, id ClientId) error {
			return nil
		},
	}
}

type defaultTopicManager struct {
	mu        sync.RWMutex
	clientMap map[ClientId]ClientItem // 客户端map
	plainMap  map[string]TopicInterface
	matchMap  map[string]TopicInterface
	handle    SubscribeHandle
}

func (m *defaultTopicManager) SetSubscribeHandle(h SubscribeHandle) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if h != nil {
		m.handle = h
	}
}
func (m *defaultTopicManager) ClientSubscribe(id ClientId, titles []string, levels ...SubscribeLevel) []SubscribeResultItem {
	m.mu.Lock()
	defer m.mu.Unlock()
	level := SubscribeQos0
	if levels != nil && len(levels) == 1 && (levels[0] == SubscribeQos2 || levels[0] == SubscribeQos1) {
		level = levels[0]
	}
	var list []SubscribeResultItem
	for _, title := range titles {
		item := SubscribeResultItem{
			SubscribeTitle: title,
			TopicTitle:     "",
			Result:         0,
			Err:            nil,
		}

		topic, err := formatTitle(title)
		if err != nil {
			item.Result = SubscribeFailed
			item.Err = err
			list = append(list, item)
			continue
		}
		topicTitle := topic.Title()
		if m.handle(topicTitle, id) != nil {
			item.Result = SubscribeFailed
			item.Err = err
			list = append(list, item)
			continue
		}
		var isAdd bool
		// 处理clientId添加到topic,topic添加到列表
		topic, isAdd = m.doNewTopicOnce(id, topic)

		if isAdd {
			topic.SetLevel(level)
		} else {
			level = topic.GetLevel()
		}
		client, ok := m.clientMap[id]
		if ok {
			if _, topicOk := client.topicMap[title]; !topicOk {
				client.topicMap[title] = topic
			}
		} else {
			client = ClientItem{
				id:       id,
				topicMap: map[string]TopicInterface{},
			}
			client.topicMap[title] = topic
			m.clientMap[id] = client
		}
		item.TopicTitle = topicTitle
		item.Result = level
		list = append(list, item)
	}
	return list
}

func (m *defaultTopicManager) ClientUnSubscribe(id ClientId, titles []string) ([]UnSubscribeResultItem, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if titles == nil || len(titles) == 0 {
		return nil, errors.New("unSubscribe titles is empty")
	}
	var list []UnSubscribeResultItem
	item, ok := m.clientMap[id]
	if !ok {
		return nil, errors.New("not found client")
	}
	if len(item.topicMap) == 0 {
		return nil, errors.New("client is not subscribe")
	}
	for _, t := range titles {
		resItem := UnSubscribeResultItem{
			Title: t,
			Err:   nil,
		}
		if topic, itemOk := item.topicMap[t]; itemOk {
			length := topic.RemoveClient(id)
			if length == 0 {
				if topic.IsMatch() {
					delete(m.matchMap, topic.Title())
				} else {
					delete(m.plainMap, topic.Title())
				}
			}
			delete(item.topicMap, t)
		} else {
			resItem.Err = errors.New(fmt.Sprintf("client is not subscribe %s", t))
		}
		list = append(list, resItem)
	}
	return list, nil
}

func (m *defaultTopicManager) ClientUnSubscribeAll(id ClientId) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if item, ok := m.clientMap[id]; ok {
		for _, topic := range item.topicMap {
			title := topic.Title()
			if topic.IsMatch() {
				if matchTopic, matchOk := m.matchMap[title]; matchOk {
					matchTopic.RemoveClient(id)
				}
			} else {
				if plainTopic, plainOk := m.plainMap[title]; plainOk {
					plainTopic.RemoveClient(id)
				}
			}
		}
		delete(m.clientMap, id)
	}
}

func (m *defaultTopicManager) GetClientSubscribe(id ClientId, start, end int) (int, []SubscribeItem, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var subscribes subs
	var length int
	if client, ok := m.clientMap[id]; ok {
		length = len(client.topicMap)
		if length < 1 {
			return length, subscribes, nil
		}
		for _, topic := range client.topicMap {
			item := SubscribeItem{
				Id:       id,
				Title:    topic.Title(),
				TimeNano: 0,
			}
			if t, topicOk := topic.GetClients()[id]; topicOk {
				item.TimeNano = t
			}
			subscribes = append(subscribes, item)
			var list []SubscribeItem
			for index, item := range subscribes {
				if index >= start && index < end {
					list = append(list, item)
				}
			}
			return length, list, nil
		}
		sort.Sort(subscribes)
	}
	return length, subscribes, errors.New(fmt.Sprintf("not found subscribes with client(%d)", id))
}

func (m *defaultTopicManager) GetPublishClientIds(publishTitle string, isCheckTopicExists ...bool) (map[ClientId]struct{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	isCheckExist := false
	if isCheckTopicExists != nil && len(isCheckTopicExists) == 1 && isCheckTopicExists[0] == true {
		isCheckExist = true
	}
	resMap := map[ClientId]struct{}{}
	title, err := checkPublishTopicTitle(publishTitle)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("topic(%s) error: %s", publishTitle, err.Error()))
	}
	if topic, ok := m.plainMap[title]; ok {
		for id, _ := range topic.GetClients() {
			resMap[id] = struct{}{}
		}
	} else {
		if isCheckExist {
			return nil, errors.New(fmt.Sprintf("topic(%s) is not exist", publishTitle))
		}
	}
	for _, matchTopic := range m.matchMap {
		if matchTopic.MatchTitle(title) {
			for id, _ := range matchTopic.GetClients() {
				resMap[id] = struct{}{}
			}
		}
	}
	return resMap, nil
}

func (m *defaultTopicManager) GetPlainTopics(start, end int) (int, []TopicInterface) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var topics Topics
	var length int
	length = len(m.plainMap)
	if length > 0 {
		for _, topic := range m.plainMap {
			topics = append(topics, topic)
			sort.Sort(topics)
			var list []TopicInterface
			for index, t := range topics {
				if index >= start && index < end {
					list = append(list, t)
				}
			}
			return length, list
		}
	}
	return length, topics
}

func (m *defaultTopicManager) GetMatchTopics(start, end int) (int, []TopicInterface) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var topics Topics
	var length int
	length = len(m.matchMap)
	if length > 0 {
		for _, topic := range m.matchMap {
			topics = append(topics, topic)
			sort.Sort(topics)
			var list []TopicInterface
			for index, t := range topics {
				if index >= start && index < end {
					list = append(list, t)
				}
			}
			return length, list
		}
	}
	return length, topics
}
func (m *defaultTopicManager) GetOnceTopicSubscribes(title string, start, end int) (int, []SubscribeItem, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var length int
	var subscribes subs
	//title, err := checkPublishTopicTitle(t)
	//if err != nil {
	//	fmt.Println("-----  ", title)
	//	return length, subscribes, err
	//}
	var topic TopicInterface
	if plain, plainOk := m.plainMap[title]; plainOk {
		topic = plain
	} else {
		if match, matchOk := m.matchMap[title]; matchOk {
			topic = match
		} else {
			return length, subscribes, errors.New(fmt.Sprintf("not found topic with(%s)", title))
		}
	}
	ids := topic.GetClients()
	topicTitle := topic.Title()
	length = len(ids)
	if length > 0 {
		var list []SubscribeItem
		for id, t := range ids {
			subscribes = append(subscribes, SubscribeItem{
				Id:       id,
				Title:    topicTitle,
				TimeNano: t,
			})
		}
		sort.Sort(subscribes)
		for index, item := range subscribes {
			if index >= start && index < end {
				list = append(list, item)
			}
		}
		return length, list, nil
	}
	return length, subscribes, nil
}
func (m *defaultTopicManager) doNewTopicOnce(id ClientId, newTopic TopicInterface) (TopicInterface, bool) {
	var resTopic TopicInterface
	var isAdd bool
	title := newTopic.Title()
	if newTopic.IsMatch() {
		if t, ok := m.matchMap[title]; ok {
			t.AddClient(id)
			resTopic = t
		} else {
			isAdd = true
			newTopic.AddClient(id)
			m.matchMap[title] = newTopic
			resTopic = newTopic
			//// 如果是通配的topic,校验所有普通topic,
			//for plainTitle, plainTopic := range m.plainMap {
			//	if resTopic.MatchTitle(plainTitle) {
			//		plainTopic.AddClient(id)
			//	}
			//}
		}
	} else {
		if t, ok := m.plainMap[title]; ok {
			t.AddClient(id)
			resTopic = t
		} else {
			isAdd = true
			newTopic.AddClient(id)
			m.plainMap[title] = newTopic
			resTopic = newTopic
			//// 如果不是通配的topic,校验所有通配,如果匹配,加入通配的topic的clientId
			//for _, matchTopic := range m.matchMap {
			//	if matchTopic.MatchTitle(title) {
			//		for clientId, _ := range matchTopic.GetClients() {
			//			resTopic.AddClient(clientId)
			//		}
			//	}
			//}
		}
	}
	return resTopic, isAdd
}

type subs []SubscribeItem

func (s subs) Len() int {
	return len(s)
}

func (s subs) Less(i, j int) bool {
	return s[i].Id > s[j].Id
}

func (s subs) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
