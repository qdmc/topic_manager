package topic_tree

type TopicTreeInterface interface {
	Subscribe(topic string, clientId int64) bool
	UnSubscribe(topic string, clientId int64) bool
	GetClientIdsWithTopic(topic string) []int64
	GetTopics() []string
}
