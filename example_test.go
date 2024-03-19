package topic_manager

import (
	"fmt"
	"testing"
	"time"
)

func Test_subscribe(t *testing.T) {
	var err error
	var length int
	manager := NewTopicManager()
	clients := makeClient(20000)
	start := time.Now().UnixNano()
	for _, client := range clients {
		manager.ClientSubscribe(client.id, client.topics)
	}
	end := time.Now().UnixNano()
	length, _, err = manager.GetOnceTopicSubscribes("/#", 0, 3)
	if err != nil {
		t.Fatal("err: ", err.Error())
	}
	fmt.Println("client.Length: ", length)
	ut := end - start
	fmt.Println(ut, "   ", time.Unix(0, ut))
	fmt.Println("-----------------------")
	start1 := time.Now().UnixNano()
	m, err := manager.GetPublishClientIds("/test/01")
	end1 := time.Now().UnixNano()
	if err != nil {
		t.Fatal("GetPublishClientIds.err: ", err.Error())
	}
	ut1 := end1 - start1
	fmt.Println("PublishClientIds.len: ", len(m))
	fmt.Println(ut1, "   ", time.Unix(0, ut1))
	i, _ := manager.GetMatchTopics(0, 3)
	fmt.Println("matchTopicLen: ", i)
}

func Test_unSubscribe(t *testing.T) {

}

type testClient struct {
	id     int64
	topics []string
}

func makeClient(l uint) []testClient {
	if l < 5 {
		l = 5
	}
	if l > 40000 {
		l = 40000
	}
	length := int64(l)
	var list []testClient
	for i := int64(1); i <= length; i++ {
		list = append(list, testClient{
			id:     i,
			topics: makeTopic(i),
		})
	}
	return list
}

func makeTopic(id int64) []string {
	return []string{
		"/#",
		"/+",
		fmt.Sprintf("/test/%d", id),
		fmt.Sprintf("/test/+/test_topic/%d", id),
	}
}
