package topic_tree

//func GetTopicTargets(t *topic) []string {
//	var list []string
//	if t.children != nil && len(t.children) > 0 {
//		for key, t := range t.children {
//			if key == "+" || key == "*" {
//				continue
//			}
//			childList := GetTopicTargets(t)
//			list = append(list, childList...)
//		}
//	}
//	return list
//}

type TopicTreeInterface interface {
	Subscribe(topic string, clientId int64) bool
	UnSubscribe(topic string, clientId int64) bool
	GetClientIdsWithTopic(topic string) []int64
	GetTopics() []string
}
