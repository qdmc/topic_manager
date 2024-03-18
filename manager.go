package topic_manager

type SubscribeResultItem struct {
	Title  string
	Result SubscribeResult
	Err    error
}

type UnSubscribeResultItem struct {
	Title string
	Err   error
}

// TopicManagerInterface      主题管理器通用接口
type TopicManagerInterface interface {
	ClientSubscribe(id ClientId, titles []string) []SubscribeResultItem     // 客户端订阅
	ClientUnSubscribe(id ClientId, titles []string) []UnSubscribeResultItem // 客户端取消订阅
	ClientUnSubscribeAll(id ClientId)                                       // 客户端取消所有订阅
	GetClientSubscribe() []TopicInterface                                   // 客户端订阅列表
	GetPublishClientIds(title string) (map[ClientId]struct{}, error)        // 消息发布受众列表
	GetPlainTopics() []TopicInterface                                       // 普通topic列表
	GetMatchTopics() []TopicInterface                                       // 匹配topic列表
}
