package topic_manager

// SubscribeLevel    订阅级别及订阅返回
type SubscribeLevel = uint8

const (
	SubscribeQos0   SubscribeLevel = 0x00
	SubscribeQos1   SubscribeLevel = 0x01
	SubscribeQos2   SubscribeLevel = 0x02
	SubscribeFailed SubscribeLevel = 0x80
)

type ClientId = int64

// LayerSeparation      主题title分隔符,开头与分层用
const LayerSeparation = "/"
