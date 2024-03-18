package topic_manager

type SubscribeResult = uint8

const (
	SubscribeQos0   SubscribeResult = 0x00
	SubscribeQos1   SubscribeResult = 0x01
	SubscribeQos2   SubscribeResult = 0x02
	SubscribeFailed SubscribeResult = 0x80
)

type ClientId = int64

// LayerSeparation      主题title分隔符,开头与分层用
const LayerSeparation = "/"
