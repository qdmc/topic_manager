package topic_manager

import (
	"regexp"
	"time"
)

// TopicInterface       主题通用接口
type TopicInterface interface {
	IsMatch() bool                  // 是否是通配的topic
	MatchTitle(title string) bool   // 正则title,通配的topic有效
	Title() string                  // 返回主题名
	AddClient(id ClientId)          // 增加一个客户端
	RemoveClient(id ClientId) int   // 主题删除一个客户端,返加剩下的客户端总数
	GetClients() map[ClientId]int64 // 返回客户端列表
	GetLevel() SubscribeLevel
	SetLevel(l SubscribeLevel)
	GetCreateNano() int64
}

// commonTopic 普通主题结构
type defaultTopic struct {
	title      string
	targets    []string
	match      *regexp.Regexp
	isMatch    bool
	clientMap  map[ClientId]int64
	level      SubscribeLevel
	createNano int64
}

func newCommonTopic(f *titleFormat) *defaultTopic {
	return &defaultTopic{
		title:      f.title,
		targets:    f.targets,
		isMatch:    f.isMatch,
		match:      f.match,
		level:      SubscribeQos0,
		clientMap:  map[ClientId]int64{},
		createNano: time.Now().UnixNano(),
	}
}
func (t *defaultTopic) GetCreateNano() int64 {
	return t.createNano
}
func (t *defaultTopic) SetLevel(l SubscribeLevel) {
	if l == SubscribeQos0 || l == SubscribeQos2 || l == SubscribeQos1 {
		t.level = l
	}
}

func (t *defaultTopic) GetLevel() SubscribeLevel {
	return t.level
}
func (t *defaultTopic) MatchTitle(title string) bool {
	if t.isMatch {
		return t.match.MatchString(title)
	} else {
		return t.title == title
	}
}

func (t *defaultTopic) IsMatch() bool {
	return t.isMatch
}

// Title              返回主题名
func (t *defaultTopic) Title() string {
	return t.title
}

// Targets            主题名拆分后的列表
func (t *defaultTopic) Targets() []string {
	return t.targets
}

// GetClients         客户端列表
func (t *defaultTopic) GetClients() map[ClientId]int64 {
	return t.clientMap
}

// AddClient         主题新增一个客户端
func (t *defaultTopic) AddClient(id ClientId) {
	t.clientMap[id] = time.Now().UnixNano()
}

// RemoveClient     主题删除一个客户端,返加剩下的客户端总数
func (t *defaultTopic) RemoveClient(id ClientId) int {
	delete(t.clientMap, id)
	return len(t.clientMap)
}

type Topics []TopicInterface

func (ts Topics) Len() int {
	return len(ts)
}

func (ts Topics) Less(i, j int) bool {
	return ts[i].Title() > ts[j].Title()
}

func (ts Topics) Swap(i, j int) {
	ts[i], ts[j] = ts[j], ts[i]
}
