package topic_manager

import "regexp"

// TopicInterface       主题通用接口
type TopicInterface interface {
	IsMatch() bool                     // 是否是通配的topic
	MatchTitle(title string) bool      // 正则title,通配的topic有效
	Title() string                     // 返回主题名
	AddClient(id ClientId)             // 增加一个客户端
	RemoveClient(id ClientId) int      // 主题删除一个客户端,返加剩下的客户端总数
	GetClients() map[ClientId]struct{} // 返回客户端列表
	GetLevel() SubscribeLevel
	SetLevel(l SubscribeLevel)
}

// commonTopic 普通主题结构
type defaultTopic struct {
	title     string
	targets   []string
	match     *regexp.Regexp
	isMatch   bool
	clientMap map[ClientId]struct{}
	level     SubscribeLevel
}

func newCommonTopic(f *titleFormat) *defaultTopic {
	return &defaultTopic{
		title:     f.title,
		targets:   f.targets,
		isMatch:   f.isMatch,
		match:     f.match,
		level:     SubscribeQos0,
		clientMap: map[ClientId]struct{}{},
	}
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
	}
	return false
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
func (t *defaultTopic) GetClients() map[ClientId]struct{} {
	return t.clientMap
}

// AddClient         主题新增一个客户端
func (t *defaultTopic) AddClient(id ClientId) {
	t.clientMap[id] = struct{}{}
}

// RemoveClient     主题删除一个客户端,返加剩下的客户端总数
func (t *defaultTopic) RemoveClient(id ClientId) int {
	delete(t.clientMap, id)
	return len(t.clientMap)
}
